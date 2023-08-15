package binanceFuture

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
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

type Job struct {
	TaskNum      int
	TotalSymbols int
	Symbol       Symbol
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
		global.Zap.Info("创建表完成", zap.Duration("耗时", elapsedTime))
	}

	var wg sync.WaitGroup
	startTime := time.Now()
	// 初始化 mpb.Progress
	p := mpb.New(mpb.WithWaitGroup(&wg))
	jobs := make(chan Job, global.Config.Mahakala.MaxUpdateRoutine)
	for w := 1; w <= global.Config.Mahakala.MaxUpdateRoutine; w++ {
		go b.updateHistoryKLines(jobs, global.Config.Mahakala.KlineInterval, &wg, p)
	}
	for i, symbol := range b.Symbols {
		wg.Add(1)
		jobs <- Job{
			TaskNum:      i,
			TotalSymbols: totalSymbols,
			Symbol:       symbol,
		}
	}
	wg.Wait()
	p.Wait() // 等待所有的进度条完成
	close(jobs)
	elapsedTime := time.Since(startTime)
	global.Zap.Info("更新历史K线完成", zap.Duration("耗时", elapsedTime))
}

func (b *BinanceFuture) getLeverageBracket() {
	const url = "/fapi/v1/leverageBracket"
	weight := 1
	b.checkLimitWeight(weight)

	// 鉴权参数
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	query := "timestamp=" + timestamp
	signature := func(payload string) string {
		mac := hmac.New(sha256.New, []byte(b.SecretKey))
		mac.Write([]byte(query))
		signature := hex.EncodeToString(mac.Sum(nil))
		return signature
	}(query)

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

func (b *BinanceFuture) updateHistoryKLines(jobs <-chan Job, interval string, wg *sync.WaitGroup, p *mpb.Progress) {
	for job := range jobs {
		func(symbolInfo Symbol, taskNum, totalSymbols int) {
			defer wg.Done()
			symbol := symbolInfo.Symbol

			// 获取数据库中最新的一条 K 线数据的时间
			var lastTime sql.NullTime
			err := b.DB.Raw(`SELECT MAX(time) FROM "` + symbol + `"`).Scan(&lastTime).Error
			if err != nil {
				global.Zap.Error(fmt.Sprintf("从数据库获取 %s 的最后时间失败:", symbol), zap.Error(err))
				return
			}

			// 如果没有数据，那么 startTime 为该交易对的上线时间，否则为最新数据的时间
			startTime := symbolInfo.OnboardDate
			if lastTime.Valid {
				startTime = lastTime.Time.UnixNano() / 1e6
			}

			formatTime := func(fromTimestamp, progress int64) string {
				return time.Unix(0, (fromTimestamp+progress)*int64(time.Millisecond)).Format("2006-01-02 15:04")
			}
			timeDecorator := decor.Any(func(s decor.Statistics) string {
				elapsed := formatTime(startTime, s.Current)
				return fmt.Sprintf("已更新至%s", elapsed)
			})
			bar := p.AddBar(0,
				mpb.BarOptional(mpb.BarRemoveOnComplete(), true),
				mpb.PrependDecorators(
					decor.Name(fmt.Sprintf("(%d/%d):%s", taskNum+1, totalSymbols, symbol)),
					timeDecorator,
				),
				mpb.AppendDecorators(
					decor.Percentage(),
				),
			)

			client := resty.New()
			client.SetRetryCount(3).SetRetryWaitTime(5 * time.Second).SetRetryMaxWaitTime(20 * time.Second)

			for {
				getLastKlineTime := startTime
				// 获取数据库中最新的一条 K 线数据的时间
				var lastTime sql.NullTime
				err := b.DB.Raw(`SELECT MAX(time) FROM "` + symbol + `"`).Scan(&lastTime).Error
				if err != nil {
					global.Zap.Error(fmt.Sprintf("从数据库获取 %s 的最后时间失败:", symbol), zap.Error(err))
					current := time.Now().UnixNano() / 1e6
					current = current - current%60000
					totalProgress := current - startTime
					bar.SetTotal(totalProgress, true)
					bar.SetCurrent(totalProgress)
					return
				}

				// 如果没有数据，那么 startTime 为该交易对的上线时间，否则为最新数据的时间
				if lastTime.Valid {
					getLastKlineTime = lastTime.Time.UnixNano() / 1e6
				}

				// 获取新的 K 线数据
				const klinesUrl = "/fapi/v1/klines"
				weight := 5
				b.checkLimitWeight(weight)

				var klines [][]interface{}
				resp, err := client.R().
					SetQueryParams(map[string]string{
						"symbol":    symbol,
						"interval":  interval,
						"startTime": strconv.FormatInt(getLastKlineTime, 10),
						"limit":     "1000",
					}).
					SetResult(&klines).
					Get(b.BaseUrl + klinesUrl)

				if err != nil || resp.StatusCode() != 200 {
					global.Zap.Error(fmt.Sprintf("获取 %s 的K线数据失败:", symbol), zap.Error(err))
					current := time.Now().UnixNano() / 1e6
					current = current - current%60000
					totalProgress := current - startTime
					bar.SetTotal(totalProgress, true)
					bar.SetCurrent(totalProgress)
					return
				}

				// 更新数据库
				for _, kline := range klines {
					kTime := time.Unix(int64(kline[0].(float64)/1000), 0)
					kOpen := kline[1].(string)
					kHigh := kline[2].(string)
					kLow := kline[3].(string)
					kClose := kline[4].(string)
					kVolume := kline[5].(string)

					err = b.DB.Exec(`
                INSERT INTO "`+symbol+`" (time, open, high, low, close, volume)
                VALUES (?, ?, ?, ?, ?, ?)
                ON CONFLICT (time) DO UPDATE
                SET open = ?, high = ?, low = ?, close = ?, volume = ?;
            `, kTime, kOpen, kHigh, kLow, kClose, kVolume, kOpen, kHigh, kLow, kClose, kVolume).Error

					if err != nil {
						global.Zap.Error(fmt.Sprintf("更新 %s 的K线数据到数据库失败:", symbol), zap.Error(err))
						current := time.Now().UnixNano() / 1e6
						current = current - current%60000
						totalProgress := current - startTime
						bar.SetTotal(totalProgress, true)
						bar.SetCurrent(totalProgress)

						return
					}
				}

				// 更新进度条
				current := time.Now().UnixNano() / 1e6
				current = current - current%60000
				totalProgress := current - startTime
				bar.SetTotal(totalProgress, false)
				bar.SetCurrent(int64(klines[len(klines)-1][0].(float64)) - startTime)

				// 如果获取的 K 线数据时间已经接近现在，跳出循环
				lastKlineTime := time.Unix(int64(klines[len(klines)-1][0].(float64)/1000), 0)
				if time.Since(lastKlineTime) <= Interval[global.Config.Mahakala.KlineInterval] {
					bar.SetTotal(totalProgress, true)
					bar.SetCurrent(totalProgress)
					break
				}
			}
		}(job.Symbol, job.TaskNum, job.TotalSymbols)
	}
}

func (b *BinanceFuture) checkLimitWeight(weight int) {
	for {
		b.LimitLock.Lock()
		if b.LimitWeight > weight {
			b.LimitWeight -= weight
			b.LimitLock.Unlock()
			go func() {
				time.Sleep(time.Minute * 1)
				b.LimitLock.Lock()
				b.LimitWeight += weight
				b.LimitLock.Unlock()
			}()
			break
		} else {
			b.LimitLock.Unlock()
			time.Sleep(time.Second * 1)
		}
	}
}
