package db

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xormplus/xorm"
	"log"
	"time"
)

var DbEngine *xorm.Engine

func InitDb() {
	driverName := beego.AppConfig.String("db_connection")
	dbHost := beego.AppConfig.String("db_host")
	dbPort := beego.AppConfig.String("db_port")
	dbDatabase := beego.AppConfig.String("db_database")
	dbUserName := beego.AppConfig.String("db_username")
	dbPassword := beego.AppConfig.String("db_password")
	var dbMaxOpenConns, dbIdleConns int
	var maxErr, idleErr error
	if dbMaxOpenConns, maxErr = beego.AppConfig.Int("db_max_open_conns"); maxErr != nil {
		dbMaxOpenConns = 30
	}
	if dbIdleConns, idleErr = beego.AppConfig.Int("db_max_idle_conns"); idleErr != nil {
		dbIdleConns = 5
	}
	logs.Info("数据库连接信息： driverName: %s, dbHost: %s, dbPort: %s, dbDatabase: %s, dbUserName: %s, dbMaxOpenConns: %d, dbMaxIdleConns: %d", driverName, dbHost, dbPort, dbDatabase, dbUserName, dbMaxOpenConns, dbIdleConns)

	DsName := dbUserName + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbDatabase + "?charset=utf8mb4"
	err := errors.New("")
	DbEngine, err = xorm.NewEngine(driverName, DsName)
	if err != nil && err.Error() != "" {
		log.Fatal(err.Error())
	}
	/*err = DbEngine.RegisterSqlMap(xorm.Xml("./db/sql/xml", ".xml"))
	if err != nil {
		log.Fatalf("db error: %#v\n", err.Error())
	}*/
	if isShowSql, errs := beego.AppConfig.Bool("db_is_show_sql"); errs == nil {
		DbEngine.ShowSQL(isShowSql)
	}
	DbEngine.SetMaxOpenConns(dbMaxOpenConns)
	DbEngine.SetMaxIdleConns(dbIdleConns)
	DbEngine.SetConnMaxLifetime(time.Duration(6) * time.Hour)

	logs.Info("init %s database ok", driverName)
}
