package initial

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"time"
)

const (
	Postgresql = "postgres"
)

// interface of different databases
type database interface {
	// Name returns the name of database
	dbName() string
	// DB Connect test
	ping(timeout, interval int) error
	// Register registers the database which will be used
	register(alias ...string) error
}

func InitDb() {
	var db database
	db = newPGSQL()
	//connect
	if err := db.ping(60, 2); err != nil {
		logs.Error("ping database err. %v", err)
		return
	}
	if err := db.register(); err != nil {
		logs.Error("register database err. %v", err)
		return
	}
	//config
	sqlDB, err := orm.GetDB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetConnMaxLifetime(time.Duration(30) * time.Second)
}

func TruncateDB(dbType string) {
	if dbType == Postgresql {
		sql := "SELECT tablename FROM pg_tables WHERE tablename NOT LIKE 'pg%' AND tablename NOT LIKE 'sql_%' ORDER BY tablename"
		var tables []string
		if _, err := orm.NewOrm().Raw(sql).QueryRows(&tables); err != nil {
			logs.Error("get apigw tables error. err %v", err)
			return
		}
		for _, value := range tables {
			_, _ = orm.NewOrm().Raw("TRUNCATE " + value).Exec()
		}
	}
}
