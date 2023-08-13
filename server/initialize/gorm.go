package initialize

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"server/global"
)

var defaultDBName = "mahakala"

func Gorm() {
	SetInitialDB()
	SetDefaultDB()
}

func SetInitialDB() {
	// 设置初始数据库
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s %s",
		global.Config.DB.Host,
		global.Config.DB.Port,
		global.Config.DB.User,
		global.Config.DB.Password,
		global.Config.DB.Dbname,
		global.Config.DB.Config)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		global.Zap.Error("Error connecting to database:", zap.Error(err))
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		global.Zap.Error("Error getting SQL DB object:", zap.Error(err))
		panic(err)
	}

	// 检查数据库是否存在
	var count int
	query := fmt.Sprintf("SELECT COUNT(datname) FROM pg_catalog.pg_database WHERE datname = '%s'", defaultDBName)
	err = sqlDB.QueryRow(query).Scan(&count)
	if err != nil {
		global.Zap.Error("Error checking if database exists:", zap.Error(err))
		panic(err)
	}

	// 创建数据库
	if count == 0 {
		_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s;", defaultDBName))
		if err != nil {
			global.Zap.Error(fmt.Sprintf("Error creating  database %s:", defaultDBName), zap.Error(err))
			panic(err)
		}
		global.Zap.Info(fmt.Sprintf("Database %s created successfully!", defaultDBName))
	} else {
		global.Zap.Info(fmt.Sprintf("Database %s already exists!", defaultDBName))
	}

	global.InitialDB = db
}

func SetDefaultDB() {
	// 设置默认数据库
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s %s",
		global.Config.DB.Host,
		global.Config.DB.Port,
		global.Config.DB.User,
		global.Config.DB.Password,
		defaultDBName,
		global.Config.DB.Config)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		global.Zap.Error("Error connecting to database:", zap.Error(err))
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		global.Zap.Error("Error getting SQL DB object:", zap.Error(err))
		panic(err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	global.DB = db
}
