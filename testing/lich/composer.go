package lich

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

var (
	retry    int
	noDown   bool
	yamlPath string
	pathHash string
	services map[string]*Container
)

func init() {
	flag.StringVar(&yamlPath, "f", "docker-compose.yaml", "composer yaml path.")
	flag.IntVar(&retry, "retry", 5, "number of retries on network failure.")
	flag.BoolVar(&noDown, "nodown", false, "containers are not recycled.")
}

func runCompose(args ...string) (output []byte, err error) {
	if _, err = os.Stat(yamlPath); os.IsNotExist(err) {
		log.Errorf("os.Stat(%s) composer yaml is not exist!", yamlPath)
		return
	}
	if yamlPath, err = filepath.Abs(yamlPath); err != nil {
		log.Errorf("filepath.Abs(%s) error(%v)", yamlPath, err)
		return
	}
	pathHash = fmt.Sprintf("%x", md5.Sum([]byte(yamlPath)))[:9]
	args = append([]string{"-f", yamlPath, "-p", pathHash}, args...)
	if output, err = exec.Command("docker-compose", args...).CombinedOutput(); err != nil {
		log.Errorf("exec.Command(docker-compose) args(%v) stdout(%s) error(%v)", args, string(output), err)
		return
	}
	return
}

// Setup setup UT related environment dependence for everything.
func Setup() (err error) {
	if _, err = runCompose("up", "-d"); err != nil {
		return
	}
	defer func() {
		if err != nil {
			err = Teardown()
		}
	}()
	if _, err = getServices(); err != nil {
		return
	}
	_, err = checkServices()
	return
}

// Teardown unsetup all environment dependence.
func Teardown() (err error) {
	if !noDown {
		_, err = runCompose("down")
	}
	return
}

func getServices() (output []byte, err error) {
	if output, err = runCompose("config", "--services"); err != nil {
		return
	}
	eol := []byte("\n")
	output = bytes.TrimSpace(output)
	if runtime.GOOS == "windows" {
		if bytes.Contains(output, []byte("\r\n")) {
			eol = []byte("\r\n")
		}
	}
	services = make(map[string]*Container)
	for _, svr := range bytes.Split(output, eol) {
		if output, err = runCompose("ps", "-q", string(svr)); err != nil {
			return
		}
		var (
			id   = string(bytes.TrimSpace(output))
			args = []string{"inspect", id, "--format", "'{{json .}}'"}
		)
		if output, err = exec.Command("docker", args...).CombinedOutput(); err != nil {
			log.Errorf("exec.Command(docker) args(%v) stdout(%s) error(%v)", args, string(output), err)
			return
		}
		if output = bytes.TrimSpace(output); bytes.Equal(output, []byte("")) {
			err = fmt.Errorf("service: %s | container: %s fails to launch", svr, id)
			log.Errorf("exec.Command(docker) args(%v) error(%v)", args, err)
			return
		}
		c := &Container{}
		if err = json.Unmarshal(bytes.Trim(output, "'"), c); err != nil {
			log.Errorf("json.Unmarshal(%s) error(%v)", string(output), err)
			return
		}
		services[string(svr)] = c
	}
	return
}

func checkServices() (output []byte, err error) {
	defer func() {
		if err != nil && retry > 0 {
			retry--
			time.Sleep(time.Second * 5)
			_, _ = getServices()
			output, err = checkServices()
			return
		}
	}()
	for svr, c := range services {
		if err = c.Healthcheck(); err != nil {
			log.Errorf("healthcheck(%s) error(%v) retrying %d times...", svr, err, retry)
			return
		}
	}
	return
}

func RemoveLocalImagesBySuffix(suffixes []string) error {
	output, err := runCompose("config", "--images")
	if err != nil {
		return err
	}
	eol := "\n"
	imagesStr := string(bytes.TrimSpace(output))
	if runtime.GOOS == "windows" {
		if strings.Contains(imagesStr, "\r\n") {
			eol = "\r\n"
		}
	}
	for _, img := range strings.Split(imagesStr, eol) {
		for _, s := range suffixes {
			if strings.HasSuffix(img, s) {
				res, re := exec.Command("docker", "rmi", img).CombinedOutput()
				if re != nil {
					if !bytes.Contains(res, []byte("No such image")) {
						return fmt.Errorf(
							"exec.Command(docker, rmi, %s) stdout(%s) error(%v)",
							img, string(res), re,
						)
					}
					log.Infof(
						"exec.Command(docker, rmi, %s) stdout(%s) error(%v)",
						img, string(res), re,
					)
					img = strings.Replace(img, "_", "-", 1)
					res, re = exec.Command("docker", "rmi", img).CombinedOutput()
					if re != nil {
						return fmt.Errorf(
							"exec.Command(docker, rmi, %s) stdout(%s) error(%v)",
							img, string(res), re,
						)
					}
				}
				break
			}
		}
	}
	return nil
}
