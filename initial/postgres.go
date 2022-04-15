/*
@Time : 20-8-11
@Author : jzd
@Project: apigw
*/
package initial

import (
	"bee-micro/config"
	"database/sql"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // pgsql driver
)

type postgreSQL struct {
	host     string
	port     string
	username string
	password string
	database string
	sslmode  string
	timezone string
	//todo other config
}

// Name returns the db name
func (p *postgreSQL) dbName() string {
	return p.database
}

// Ping ...
func (p *postgreSQL) ping(timeout, interval int) error {
	dataSourceName := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s timezone=%s sslmode=%s",
		p.username, p.password, p.database, p.host, p.port, p.timezone, p.sslmode)
	db, err := sql.Open(Postgresql, dataSourceName)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			// postgres error Invalid Catalog Name;
			// See http://www.postgresql.org/docs/current/static/protocol-error-fields.html for details of the fields
			if e.Code == "3D000" {
				dataSource := fmt.Sprintf("user=%s password=%s host=%s port=%s timezone=%s sslmode=%s",
					p.username, p.password, p.host, p.port, p.timezone, p.sslmode)
				dbForCreateDatabase, err := sql.Open(Postgresql, dataSource)
				if err != nil {
					return err
				}
				defer dbForCreateDatabase.Close()
				_, err = dbForCreateDatabase.Exec(fmt.Sprintf("CREATE DATABASE %s;", p.database))
				if err != nil {
					return err
				}
			} else {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

// Register
func (p *postgreSQL) register(alias ...string) error {
	if err := orm.RegisterDriver(Postgresql, orm.DRPostgres); err != nil {
		return err
	}
	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	info := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s timezone=%s sslmode=%s",
		p.host, p.port, p.username, p.password, p.database, p.timezone, p.sslmode)

	return orm.RegisterDataBase(an, Postgresql, info)
}

// newPGSQL returns an instance of postgres
func newPGSQL() database {

	conf, err := config.GetDB()
	if err != nil {
		logs.Error("get database from config center err, %v", err.Error())
		return nil
	}
	return &postgreSQL{
		host:     conf.Host,
		port:     conf.Port,
		username: conf.User,
		password: conf.Password,
		database: conf.Name,
		timezone: conf.Timezone,
		sslmode:  "disable", //default
	}
}
