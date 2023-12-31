package exchange

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"server/global"
	"server/model/response"
)

type Exchange interface {
	Init()
	InitExchangeInfo()
	UpdateExchangeInfo()
	UpdateKlinesWithProgress()
	UpdateKlines()
	GetName() string
	GetSymbols() []string
	CheckSymbol(symbol string) bool
	GetSymbolInfo(symbol string) map[string]any
	GetKlines(symbol, interval string) (klines []response.Kline, err error)
}

type BaseExchange struct {
	Name    string   `json:"name"`
	Alias   string   `json:"alias"`
	BaseUrl string   `json:"baseUrl"`
	Enabled bool     `json:"-"`
	DB      *gorm.DB `json:"-"`
}

var Exchanges map[string]Exchange

func CreateDataBase(name string) {
	// 检查数据库是否存在
	var count int
	err := global.InitialDB.Raw("SELECT COUNT(datname) FROM pg_catalog.pg_database WHERE datname = ?", name).Scan(&count).Error
	if err != nil {
		global.Zap.Error("检查数据库是否存在时出错:", zap.Error(err))
		panic(err)
	}

	// 创建数据库
	if count == 0 {
		createDBSQL := fmt.Sprintf("CREATE DATABASE %s;", name)
		err = global.InitialDB.Exec(createDBSQL).Error
		if err != nil {
			global.Zap.Error(fmt.Sprintf("创建数据库 %s 时出错:", name), zap.Error(err))
			panic(err)
		}
		global.Zap.Info(fmt.Sprintf("数据库 %s 创建成功！", name))
	} else {
		global.Zap.Info(fmt.Sprintf("数据库 %s 已存在！", name))
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		global.Zap.Error("连接数据库时出错:", zap.Error(err))
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		global.Zap.Error("获取 SQL DB 对象时出错:", zap.Error(err))
		panic(err)
	}

	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(500)

	// 创建timescaledb扩展
	err = db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb;").Error
	if err != nil {
		global.Zap.Error(fmt.Sprintf("在数据库 %s 上创建 timescaledb 扩展时出错:", name), zap.Error(err))
		panic(err)
	}

	var exists bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb')").Scan(&exists).Error
	if err != nil {
		global.Zap.Error(fmt.Sprintf("检查数据库 %s 上 timescaledb 扩展是否存在时出错:", name), zap.Error(err))
		panic(err)
	}

	if exists {
		global.Zap.Info(fmt.Sprintf("数据库 %s 上已存在 timescaledb 扩展！", name))
	} else {
		global.Zap.Error(fmt.Sprintf("在数据库 %s 上创建 timescaledb 扩展失败！", name))
		panic(err)
	}

	return db
}

func UpdateKlines() {
	if global.Config.Mahakala.UpdateKline {
		for _, ex := range Exchanges {
			ex.UpdateExchangeInfo()
			go ex.UpdateKlines()
		}
	}
}

func GetExchanges() []string {
	var exchanges []string
	for _, ex := range Exchanges {
		exchanges = append(exchanges, ex.GetName())
	}
	return exchanges
}
