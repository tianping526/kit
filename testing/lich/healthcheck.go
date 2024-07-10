package lich

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	// Register go-sql-driver stuff
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var healthchecks = map[string]func(*Container) error{"mysql": checkMysql, "mariadb": checkMysql, "postgres": checkPg}

// Healthcheck check container health.
func (c *Container) Healthcheck() (err error) {
	if status, health := c.State.Status, c.State.Health.Status; !c.State.Running || (health != "" && health != "healthy") {
		err = fmt.Errorf("service: %s | container: %s not running", c.GetImage(), c.GetID())
		log.Errorf("docker status(%s) health(%s) error(%v)", status, health, err)
		return
	}
	if check, ok := healthchecks[c.GetImage()]; ok {
		err = check(c)
		return
	}
	for proto, ports := range c.NetworkSettings.Ports {
		if id := c.GetID(); !strings.Contains(proto, "tcp") {
			log.Errorf("container: %s proto(%s) unsupported.", id, proto)
			continue
		}
		for _, publish := range ports {
			var (
				ip      = net.ParseIP(publish.HostIP)
				port, _ = strconv.Atoi(publish.HostPort)
				tcpAddr = &net.TCPAddr{IP: ip, Port: port}
				tcpConn *net.TCPConn
			)
			if tcpConn, err = net.DialTCP("tcp", nil, tcpAddr); err != nil {
				log.Errorf("net.DialTCP(%s:%s) error(%v)", publish.HostIP, publish.HostPort, err)
				return
			}
			err = tcpConn.Close()
		}
	}
	return
}

func checkMysql(c *Container) (err error) {
	var ip, port, user, passwd string
	for _, env := range c.Config.Env {
		splits := strings.Split(env, "=")
		if strings.Contains(splits[0], "MYSQL_ROOT_PASSWORD") {
			user, passwd = "root", splits[1]
			continue
		}
		if strings.Contains(splits[0], "MYSQL_ALLOW_EMPTY_PASSWORD") {
			user, passwd = "root", ""
			continue
		}
		if strings.Contains(splits[0], "MYSQL_USER") {
			user = splits[1]
			continue
		}
		if strings.Contains(splits[0], "MYSQL_PASSWORD") {
			passwd = splits[1]
			continue
		}
	}
	var db *sql.DB
	if ports, ok := c.NetworkSettings.Ports["3306/tcp"]; ok {
		ip, port = ports[0].HostIP, ports[0].HostPort
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, passwd, ip, port)
	if db, err = sql.Open("mysql", dsn); err != nil {
		log.Errorf("sql.Open(mysql) dsn(%s) error(%v)", dsn, err)
		return
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err = db.Ping(); err != nil {
		log.Errorf("ping(db) dsn(%s) error(%v)", dsn, err)
	}
	return
}

func checkPg(c *Container) (err error) {
	var ip, port, user, passwd, dbName string
	user = "postgres"
	for _, env := range c.Config.Env {
		splits := strings.Split(env, "=")
		if strings.Contains(splits[0], "POSTGRES_PASSWORD") {
			passwd = splits[1]
			continue
		}
		if strings.Contains(splits[0], "POSTGRES_USER") {
			user = splits[1]
			continue
		}
		if strings.Contains(splits[0], "POSTGRES_DB") {
			dbName = splits[1]
			continue
		}
	}
	if dbName == "" {
		dbName = user
	}
	var db *sql.DB
	if ports, ok := c.NetworkSettings.Ports["5432/tcp"]; ok {
		ip, port = ports[0].HostIP, ports[0].HostPort
	}
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, passwd, ip, port, dbName)
	if db, err = sql.Open("pgx", dsn); err != nil {
		log.Errorf("sql.Open(pgx) dsn(%s) error(%v)", dsn, err)
		return
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err = db.Ping(); err != nil {
		log.Errorf("ping(db) dsn(%s) error(%v)", dsn, err)
	}
	return
}
