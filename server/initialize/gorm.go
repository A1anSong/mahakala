package initialize

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"server/global"
)

var defaultDBName = "mahakala"

func Gorm() {
	setInitialDB()
	setDefaultDB()
}

func setDefaultDB() {
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
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {
			global.Zap.Error("Error closing SQL DB connection:", zap.Error(err))
			panic(err)
		}
	}(sqlDB)

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	global.DB = db
}

func setInitialDB() {
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
	defer func(initialDB *sql.DB) {
		err := initialDB.Close()
		if err != nil {
			global.Zap.Error("Error closing SQL DB connection:", zap.Error(err))
			panic(err)
		}
	}(sqlDB)

	var count int
	query := fmt.Sprintf("SELECT COUNT(datname) FROM pg_catalog.pg_database WHERE datname = '%s'", defaultDBName)
	err = sqlDB.QueryRow(query).Scan(&count)
	if err != nil {
		global.Zap.Error("Error checking if database exists:", zap.Error(err))
		panic(err)
	}

	if count == 0 {
		_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s;", defaultDBName))
		if err != nil {
			global.Zap.Error("Error creating new database:", zap.Error(err))
			panic(err)
		}
		global.Zap.Info(fmt.Sprintf("Database %s created successfully!", defaultDBName))
	} else {
		global.Zap.Info(fmt.Sprintf("Database %s already exists!", defaultDBName))
	}

	global.InitialDB = db
}
