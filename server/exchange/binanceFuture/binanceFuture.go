package binanceFuture

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"reflect"
	"server/exchange"
	"server/global"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BinanceFuture struct {
	exchange.BaseExchange
	LimitWeight      int
	Symbols          []Symbol
	SymbolsSet       StringSet
	LeverageBrackets []LeverageBracket
	IsFirstTime      bool
	LimitLock        sync.Mutex
}

func (b *BinanceFuture) Init() {
	b.LimitWeight = 2400
	b.SymbolsSet = make(map[string]struct{})
	b.IsFirstTime = true
}

func (b *BinanceFuture) UpdateExchangeInfo() {
	const url = "/fapi/v1/exchangeInfo"
	weight := 1
	b.checkLimitWeight(weight)

	global.Zap.Info("开始获取币安交易所信息...")
	var exchangeInfo ExchangeInfo
	client := resty.New()
	resp, err := client.R().
		SetResult(&exchangeInfo).
		Get(b.BaseUrl + url)
	if err != nil {
		global.Zap.Error("从币安 API 获取数据时出错:", zap.Error(err))
		return
	}

	if resp.StatusCode() != 200 {
		global.Zap.Error("币安 API 响应状态码非 200:", zap.Int("status code", resp.StatusCode()))
		return
	}

	// Create newSymbolsSet
	newSymbolsSet := NewStringSet()
	for _, symbol := range exchangeInfo.Symbols {
		if symbol.Status == "TRADING" && strings.HasSuffix(symbol.Symbol, "USDT") {
			newSymbolsSet.Add(symbol.Symbol)
		}
	}

	shouldCreateTable := false
	// Check differences
	if !reflect.DeepEqual(b.SymbolsSet, newSymbolsSet) {
		if !b.IsFirstTime {
			added := newSymbolsSet.Difference(b.SymbolsSet)
			for symbol := range added {
				// TODO: 添加交易对提醒
				fmt.Println("添加了交易对:" + symbol)
				// Send message about added symbol
				// feishu.send_post_message(...)  // Adjust to your Go implementation
			}

			removed := b.SymbolsSet.Difference(newSymbolsSet)
			for symbol := range removed {
				// TODO: 移除交易对提醒
				fmt.Println("移除了交易对:" + symbol)
				// Send message about removed symbol
				// feishu.send_post_message(...)  // Adjust to your Go implementation
			}
		}

		shouldCreateTable = true
		// Filter symbols and get the full SymbolInfo objects
		var symbols []Symbol
		for _, symbol := range exchangeInfo.Symbols {
			if symbol.Status == "TRADING" && strings.HasSuffix(symbol.Symbol, "USDT") {
				var priceFilters []Filter
				for _, filter := range symbol.Filters {
					if filter.FilterType == "PRICE_FILTER" {
						priceFilters = append(priceFilters, filter)
					}
				}
				symbol.Filters = priceFilters
				symbols = append(symbols, symbol)
			}
		}
		b.Symbols = symbols
		b.SymbolsSet = newSymbolsSet
		b.IsFirstTime = false
	}

	totalSymbols := len(b.Symbols)
	global.Zap.Info("获取币安交易所信息成功", zap.Int("TRADING状态且标的资产为USDT的交易对数量", totalSymbols))

	b.getLeverageBracket()
	if shouldCreateTable {
		var wg sync.WaitGroup
		startTime := time.Now()
		for _, symbol := range b.Symbols {
			wg.Add(1)
			go b.createTable(symbol.Symbol, &wg)
		}
		wg.Wait()
		elapsedTime := time.Since(startTime)
		global.Zap.Info("创建表耗时", zap.Duration("耗时", elapsedTime))
	}
}

func (b *BinanceFuture) getLeverageBracket() {
	const url = "/fapi/v1/leverageBracket"
	weight := 1
	b.checkLimitWeight(weight)

	// 鉴权参数
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	query := "timestamp=" + timestamp
	signature := b.sign(query)

	// 构建完整的URL，确保signature是最后一个参数
	fullURL := fmt.Sprintf("%s%s?%s&signature=%s", b.BaseUrl, url, query, signature)

	var leverageBrackets []LeverageBracket
	client := resty.New()
	resp, err := client.R().
		SetHeader("X-MBX-APIKEY", b.ApiKey).
		SetResult(&leverageBrackets).
		Get(fullURL)
	if err != nil {
		global.Zap.Error("从币安 API 获取数据时出错:", zap.Error(err))
		return
	}

	if resp.StatusCode() != 200 {
		global.Zap.Error("币安 API 响应状态码非 200:", zap.Int("status code", resp.StatusCode()))
		return
	}

	b.LeverageBrackets = leverageBrackets
}

func (b *BinanceFuture) createTable(symbol string, wg *sync.WaitGroup) {
	defer wg.Done()

	// 检查表是否存在
	var result *string
	err := b.DB.Raw(`SELECT to_regclass(?)`, `public."`+symbol+`"`).Scan(&result).Error
	if err != nil {
		global.Zap.Error(fmt.Sprintf("检查表 %s 是否存在失败:", symbol), zap.Error(err))
		return
	}

	// 如果表不存在，则创建表
	if result == nil {
		err := b.DB.Exec(`
            CREATE TABLE "` + symbol + `" (
                time TIMESTAMPTZ NOT NULL,
                open NUMERIC NOT NULL,
                close NUMERIC NOT NULL,
                high NUMERIC NOT NULL,
                low NUMERIC NOT NULL,
                volume NUMERIC NOT NULL,
                PRIMARY KEY(time)
            );
        `).Error
		if err != nil {
			global.Zap.Error(fmt.Sprintf("创建表 %s 失败:", symbol), zap.Error(err))
			return
		}
	}

	// 检查表是否已经是超表
	var hypertableName string
	err = b.DB.Raw(`SELECT hypertable_name FROM timescaledb_information.hypertables WHERE hypertable_name = ?`, symbol).Scan(&hypertableName).Error
	if err != nil {
		global.Zap.Error(fmt.Sprintf("检查表 %s 是否是超表失败:", symbol), zap.Error(err))
		return
	}

	// 如果表不是超表，则创建超表
	if hypertableName == "" {
		err := b.DB.Exec(`SELECT create_hypertable(?, 'time')`, `"`+symbol+`"`).Error
		if err != nil {
			global.Zap.Error(fmt.Sprintf("创建超表 %s 失败:", symbol), zap.Error(err))
			return
		}
	}
}

func (b *BinanceFuture) checkLimitWeight(weight int) {
	for {
		b.LimitLock.Lock()
		if b.LimitWeight > 0 {
			b.LimitWeight -= weight
			b.LimitLock.Unlock()
			go b.recoverLimitWeight(weight)
			break
		} else {
			b.LimitLock.Unlock()
			time.Sleep(time.Second * 1)
		}
	}
}

func (b *BinanceFuture) recoverLimitWeight(weight int) {
	time.Sleep(time.Minute * 1)
	b.LimitLock.Lock()
	b.LimitWeight += weight
	b.LimitLock.Unlock()
}

func (b *BinanceFuture) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(b.SecretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
