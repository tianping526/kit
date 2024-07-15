package lich

import (
	"context"
	"database/sql"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	// Register go-sql-driver stuff
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var healthchecks = map[string]func(*Container) error{
	"mysql":    checkMysql,
	"mariadb":  checkMysql,
	"postgres": checkPg,
	"mongo":    checkMongo,
}

// Healthcheck check container health.
func (c *Container) Healthcheck() error {
	status, health := c.State.Status, c.State.Health.Status
	if !c.State.Running || (health != "" && health != "healthy") {
		return fmt.Errorf(
			"docker status(%s) health(%s) error(service: %s | container: %s not running)",
			status, health, c.GetImage(), c.GetID(),
		)
	}
	if check, ok := healthchecks[c.GetImage()]; ok {
		return check(c)
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
				err     error
			)
			if tcpConn, err = net.DialTCP("tcp", nil, tcpAddr); err != nil {
				return fmt.Errorf("net.DialTCP(%s:%s) error(%v)", publish.HostIP, publish.HostPort, err)
			}
			return tcpConn.Close()
		}
	}
	return nil
}

func checkMysql(c *Container) error {
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
	var (
		db  *sql.DB
		err error
	)
	if ports, ok := c.NetworkSettings.Ports["3306/tcp"]; ok {
		ip, port = ports[0].HostIP, ports[0].HostPort
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, passwd, ip, port)
	if db, err = sql.Open("mysql", dsn); err != nil {
		return fmt.Errorf("sql.Open(mysql) dsn(%s) error(%v)", dsn, err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err = db.Ping(); err != nil {
		return fmt.Errorf("ping(db) dsn(%s) error(%v)", dsn, err)
	}
	return nil
}

func checkPg(c *Container) error {
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
	var (
		db  *sql.DB
		err error
	)
	if ports, ok := c.NetworkSettings.Ports["5432/tcp"]; ok {
		ip, port = ports[0].HostIP, ports[0].HostPort
	}
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, passwd, ip, port, dbName)
	if db, err = sql.Open("pgx", dsn); err != nil {
		return fmt.Errorf("sql.Open(pgx) dsn(%s) error(%v)", dsn, err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err = db.Ping(); err != nil {
		return fmt.Errorf("ping(db) dsn(%s) error(%v)", dsn, err)
	}
	return nil
}

func checkMongo(c *Container) error {
	var ip, port, user, passwd string
	for _, env := range c.Config.Env {
		splits := strings.Split(env, "=")
		if strings.Contains(splits[0], "MONGO_INITDB_ROOT_USERNAME") {
			user = splits[1]
			continue
		}
		if strings.Contains(splits[0], "MONGO_INITDB_ROOT_PASSWORD") {
			passwd = splits[1]
			continue
		}
	}

	if ports, ok := c.NetworkSettings.Ports["27017/tcp"]; ok {
		ip, port = ports[0].HostIP, ports[0].HostPort
	}

	db, err := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(
			fmt.Sprintf("mongodb://%s:%s@%s:%s", user, passwd, ip, port),
		),
	)
	if err != nil {
		return fmt.Errorf("mongo.Connect error(%v)", err)
	}
	defer func(mc *mongo.Client) {
		_ = mc.Disconnect(context.Background())
	}(db)
	if err = db.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("ping(mongo) error(%v)", err)
	}
	return nil
}
