package db

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xormplus/xorm"
	"log"
)

var DbEngine *xorm.Engine

func InitDb() {
	driverName := beego.AppConfig.String("db_connection")
	dbHost := beego.AppConfig.String("db_host")
	dbPort := beego.AppConfig.String("db_port")
	dbDatabase := beego.AppConfig.String("db_database")
	dbUserName := beego.AppConfig.String("db_username")
	dbPassword := beego.AppConfig.String("db_password")

	DsName := dbUserName + ":" + dbPassword + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbDatabase + "?charset=utf8mb4"
	err := errors.New("")
	DbEngine, err = xorm.NewEngine(driverName, DsName)
	if err != nil && err.Error() != "" {
		log.Fatal(err.Error())
	}
	err = DbEngine.RegisterSqlMap(xorm.Xml("./db/sql/xml", ".xml"))
	if err != nil {
		log.Fatalf("db error: %#v\n", err.Error())
	}
	if isShowSql, errs := beego.AppConfig.Bool("db_is_show_sql"); errs == nil {
		DbEngine.ShowSQL(isShowSql)
	}
	DbEngine.SetMaxOpenConns(30)
	DbEngine.SetMaxIdleConns(5)

	logs.Info("init %s database ok", driverName)
}
