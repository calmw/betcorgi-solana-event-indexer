package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func InitMysql() {
	log.Println("Init Mysql")

	var userName = os.Getenv("MYSQL_USERNAME")
	var password = os.Getenv("MYSQL_PASSWORD")
	var host = os.Getenv("MYSQL_HOST")
	var port = os.Getenv("MYSQL_PORT")
	var database = os.Getenv("MYSQL_DATABASE")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		userName,
		password,
		host,
		port,
		database,
	)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Println(fmt.Sprintf("mysql connention error ==>  %+v", err))
		panic(fmt.Sprintf("mysql connention error ==>  %+v", err))
	}

	_ = db.Callback().Create().After("gorm:after_create").Register("after_create", After)
	_ = db.Callback().Query().After("gorm:after_query").Register("after_query", After)
	_ = db.Callback().Delete().After("gorm:after_delete").Register("after_delete", After)
	_ = db.Callback().Update().After("gorm:after_update").Register("after_update", After)
	_ = db.Callback().Row().After("gorm:row").Register("after_row", After)
	_ = db.Callback().Raw().After("gorm:raw").Register("after_raw", After)

	DB = db
	log.Println(fmt.Sprintf("mysql connented"))
}

func After(db *gorm.DB) {
	db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	//sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	//log.Debug(sql)
}
