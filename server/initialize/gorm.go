package initialize

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		global.Zap.Error("连接数据库失败:", zap.Error(err))
		panic(err)
	}

	// 检查数据库是否存在
	var count int
	err = db.Raw("SELECT COUNT(datname) FROM pg_catalog.pg_database WHERE datname = ?", defaultDBName).Scan(&count).Error
	if err != nil {
		global.Zap.Error("检查数据库是否存在时出错:", zap.Error(err))
		panic(err)
	}

	// 创建数据库
	if count == 0 {
		err = db.Exec("CREATE DATABASE " + defaultDBName + ";").Error
		if err != nil {
			global.Zap.Error(fmt.Sprintf("创建数据库 %s 时出错:", defaultDBName), zap.Error(err))
			panic(err)
		}
		global.Zap.Info(fmt.Sprintf("数据库 %s 创建成功!", defaultDBName))
	} else {
		global.Zap.Info(fmt.Sprintf("数据库 %s 已存在!", defaultDBName))
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		global.Zap.Error("连接数据库失败:", zap.Error(err))
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		global.Zap.Error("获取 SQL DB 对象失败:", zap.Error(err))
		panic(err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	global.DB = db
}
