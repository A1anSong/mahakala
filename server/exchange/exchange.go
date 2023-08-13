package exchange

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"server/global"
)

type Exchange interface {
}

type BaseExchange struct {
	Name      string
	BaseUrl   string
	ApiKey    string
	SecretKey string
	DB        *gorm.DB
}

var Exchanges []Exchange

func CreateDataBase(name string) {
	// 获取初始数据库
	sqlDB, err := global.InitialDB.DB()
	if err != nil {
		global.Zap.Error("Error getting SQL DB object:", zap.Error(err))
		panic(err)
	}

	// 检查数据库是否存在
	var count int
	query := fmt.Sprintf("SELECT COUNT(datname) FROM pg_catalog.pg_database WHERE datname = '%s'", name)
	err = sqlDB.QueryRow(query).Scan(&count)
	if err != nil {
		global.Zap.Error("Error checking if database exists:", zap.Error(err))
		panic(err)
	}

	// 创建数据库
	if count == 0 {
		_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s;", name))
		if err != nil {
			global.Zap.Error(fmt.Sprintf("Error creating  database %s:", name), zap.Error(err))
			panic(err)
		}
		global.Zap.Info(fmt.Sprintf("Database %s created successfully!", name))
	} else {
		global.Zap.Info(fmt.Sprintf("Database %s already exists!", name))
	}
}

func SetDataBase(name string) *gorm.DB {
	// 设置数据库
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s %s",
		global.Config.DB.Host,
		global.Config.DB.Port,
		global.Config.DB.User,
		global.Config.DB.Password,
		name,
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

	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(500)

	// 创建timescaledb扩展
	_, err = sqlDB.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb;")
	if err != nil {
		global.Zap.Error(fmt.Sprintf("Error creating timescaledb extension on database %s:", name), zap.Error(err))
		panic(err)
	}

	var exists bool
	err = sqlDB.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb')").Scan(&exists)
	if err != nil {
		global.Zap.Error(fmt.Sprintf("Error checking if timescaledb extension exists on %s:", name), zap.Error(err))
		panic(err)
	}

	if exists {
		global.Zap.Info(fmt.Sprintf("timescaledb extension created or already exists on %s!", name))
	} else {
		global.Zap.Error(fmt.Sprintf("Failed to create timescaledb extension on %s!", name))
		panic(err)
	}

	return db
}
