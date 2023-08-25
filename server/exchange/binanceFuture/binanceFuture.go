package binanceFuture

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-module/carbon/v2"
	"github.com/shopspring/decimal"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"reflect"
	"server/exchange"
	"server/global"
	"server/model/common"
	"server/model/response"
	"server/utils"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BinanceFuture struct {
	exchange.BaseExchange
	Symbols          []Symbol          `json:"-"`
	SymbolsSet       utils.StringSet   `json:"-"`
	LeverageBrackets []LeverageBracket `json:"-"`
	LimitWeight      int               `json:"-"`
	LimitLock        sync.Mutex        `json:"-"`
}

type Job struct {
	Symbol Symbol
}

type JobWithProgress struct {
	TaskNum      int
	TotalSymbols int
	Symbol       Symbol
}

func (b *BinanceFuture) Init() {
	b.SymbolsSet = make(map[string]struct{})
	b.LimitWeight = 2400
}

func (b *BinanceFuture) InitExchangeInfo() {
	const url = "/fapi/v1/exchangeInfo"
	weight := 1
	b.checkLimitWeight(weight)

	global.Zap.Info(fmt.Sprintf("开始获取%s交易所信息", b.Name))
	var exchangeInfo ExchangeInfo
	resp, err := global.Resty.R().
		SetResult(&exchangeInfo).
		Get(b.BaseUrl + url)
	if err != nil {
		global.Zap.Error(fmt.Sprintf("从%s API 获取数据时出错:", b.Name), zap.Error(err))
		return
	}

	if resp.StatusCode() != 200 {
		global.Zap.Error(fmt.Sprintf("%s API 响应状态码: %d", b.Name, resp.StatusCode()), zap.String("response body", string(resp.Body())))
		return
	}

	var symbols []Symbol
	symbolsSet := utils.NewStringSet()
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
			symbolsSet.Add(symbol.Symbol)
		}
	}
	b.Symbols = symbols
	b.SymbolsSet = symbolsSet

	global.Zap.Info(fmt.Sprintf("获取%s交易所信息成功", b.Name), zap.Int("TRADING状态且标的资产为USDT的交易对数量", len(b.Symbols)))

	// 获取杠杆分层标准
	b.getLeverageBracket()

	// 创建表
	var wg sync.WaitGroup
	for _, symbol := range b.Symbols {
		wg.Add(1)
		go b.createTable(symbol.Symbol, &wg)
	}
	wg.Wait()
}

func (b *BinanceFuture) UpdateExchangeInfo() {
	const url = "/fapi/v1/exchangeInfo"
	weight := 1
	b.checkLimitWeight(weight)

	var exchangeInfo ExchangeInfo
	resp, err := global.Resty.R().
		SetResult(&exchangeInfo).
		Get(b.BaseUrl + url)
	if err != nil {
		global.Zap.Error(fmt.Sprintf("从%s API 获取数据时出错:", b.Name), zap.Error(err))
		return
	}

	if resp.StatusCode() != 200 {
		global.Zap.Error(fmt.Sprintf("%s API 响应状态码: %d", b.Name, resp.StatusCode()), zap.String("response body", string(resp.Body())))
		return
	}

	var symbols []Symbol
	newSymbolsSet := utils.NewStringSet()
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
			newSymbolsSet.Add(symbol.Symbol)
		}
	}

	// 比较新旧交易对
	if !reflect.DeepEqual(b.SymbolsSet, newSymbolsSet) {
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

	b.Symbols = symbols
	b.SymbolsSet = newSymbolsSet

	// 获取杠杆分层标准
	b.getLeverageBracket()

	// 创建表
	var wg sync.WaitGroup
	for _, symbol := range b.Symbols {
		wg.Add(1)
		go b.createTable(symbol.Symbol, &wg)
	}
	wg.Wait()
}

func (b *BinanceFuture) UpdateKlinesWithProgress() {
	global.Zap.Info(fmt.Sprintf("开始更新%s交易所历史K线", b.Name))
	var wg sync.WaitGroup
	start := global.Carbon.Now()
	// 初始化 mpb.Progress
	p := mpb.New(mpb.WithWaitGroup(&wg))
	jobs := make(chan JobWithProgress, global.Config.Mahakala.MaxUpdateRoutine)
	for w := 1; w <= global.Config.Mahakala.MaxUpdateRoutine; w++ {
		go b.updateHistoryKLinesWithProgress(jobs, global.Config.Mahakala.KlineInterval, &wg, p)
	}
	for i, symbol := range b.Symbols {
		wg.Add(1)
		jobs <- JobWithProgress{
			TaskNum:      i,
			TotalSymbols: len(b.Symbols),
			Symbol:       symbol,
		}
	}
	wg.Wait()
	p.Wait() // 等待所有的进度条完成
	close(jobs)
	global.Zap.Info(fmt.Sprintf("更新%s交易所历史K线完成", b.Name), zap.String("耗时", start.DiffInString()))
}

func (b *BinanceFuture) UpdateKlines() {
	global.Zap.Info(fmt.Sprintf("开始更新%s交易所K线", b.Name))
	var wg sync.WaitGroup
	start := global.Carbon.Now()
	jobs := make(chan Job, global.Config.Mahakala.MaxUpdateRoutine)
	for w := 1; w <= global.Config.Mahakala.MaxUpdateRoutine; w++ {
		go b.updateHistoryKLines(jobs, global.Config.Mahakala.KlineInterval, &wg)
	}
	for _, symbol := range b.Symbols {
		wg.Add(1)
		jobs <- Job{
			Symbol: symbol,
		}
	}
	wg.Wait()
	close(jobs)
	global.Zap.Info(fmt.Sprintf("更新%s交易所K线完成", b.Name), zap.String("耗时", start.DiffInString()))
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

func (b *BinanceFuture) getLeverageBracket() {
	const url = "/fapi/v1/leverageBracket"
	weight := 1
	b.checkLimitWeight(weight)

	// 鉴权参数
	timestamp := strconv.FormatInt(global.Carbon.Now().TimestampMilli(), 10)
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
	resp, err := global.Resty.R().
		SetHeader("X-MBX-APIKEY", b.ApiKey).
		SetResult(&leverageBrackets).
		Get(fullURL)
	if err != nil {
		global.Zap.Error(fmt.Sprintf("从%s API 获取数据时出错:", b.Name), zap.Error(err))
		return
	}

	if resp.StatusCode() != 200 {
		global.Zap.Error(fmt.Sprintf("%s API 响应状态码: %d", b.Name, resp.StatusCode()), zap.String("response body", string(resp.Body())))
		return
	}

	b.LeverageBrackets = leverageBrackets
}

func (b *BinanceFuture) createTable(symbol string, wg *sync.WaitGroup) {
	defer wg.Done()

	table := fmt.Sprintf(`%s_%s`, global.Config.Mahakala.KlineInterval, symbol)

	// 检查表是否存在
	if !b.DB.Migrator().HasTable(table) {
		err := b.DB.Table(table).Migrator().CreateTable(&common.Kline{})
		if err != nil {
			global.Zap.Error(fmt.Sprintf("创建表 %s 失败:", symbol), zap.Error(err))
			return
		}
	}

	// 检查表是否已经是超表
	var hypertableName string
	err := b.DB.Raw(`SELECT hypertable_name FROM timescaledb_information.hypertables WHERE hypertable_name = ?`, table).Scan(&hypertableName).Error
	if err != nil {
		global.Zap.Error(fmt.Sprintf("检查表 %s 是否是超表失败:", symbol), zap.Error(err))
		return
	}

	// 如果表不是超表，则创建超表
	if hypertableName == "" {
		createHyperTableSQL := fmt.Sprintf(`SELECT create_hypertable('"%s"', 'time')`, table)
		err = b.DB.Exec(createHyperTableSQL).Error
		if err != nil {
			global.Zap.Error(fmt.Sprintf("创建超表 %s 失败:", symbol), zap.Error(err))
			return
		}
	}
}

func (b *BinanceFuture) updateHistoryKLinesWithProgress(jobs <-chan JobWithProgress, interval string, wg *sync.WaitGroup, p *mpb.Progress) {
	for job := range jobs {
		func(symbolInfo Symbol, taskNum, totalSymbols int, p *mpb.Progress) {
			defer wg.Done()

			symbol := symbolInfo.Symbol
			table := fmt.Sprintf(`%s_%s`, global.Config.Mahakala.KlineInterval, symbol)

			// 获取数据库中最新的一条 K 线数据的时间，如果没有数据，那么 startTime 为上市时间，否则为最新数据的时间
			startTime := global.Carbon.CreateFromTimestampMilli(symbolInfo.OnboardDate)
			var lastKline common.Kline
			result := b.DB.Table(table).Order("time DESC").First(&lastKline)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					global.Zap.Error(fmt.Sprintf("从数据库获取 %s 的最后时间失败:", table), zap.Error(result.Error))
					return
				}
			} else {
				startTime = lastKline.Time.Carbon
			}

			timeNow := global.Carbon.Now()
			// 如果最后一条 K 线数据的时间距离现在不足一个 K 线周期，跳过
			if startTime.DiffInMinutes(timeNow) < utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes {
				return
			}

			timeDecorator := decor.Any(func(s decor.Statistics) string {
				return fmt.Sprintf("已更新至%s", startTime.AddMinutes(int(s.Current)).Layout("2006年01月02日 15时04分"))
			})
			bar := p.AddBar((startTime.DiffInMinutes(timeNow)/utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes)*utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes,
				mpb.BarOptional(mpb.BarRemoveOnComplete(), true),
				mpb.PrependDecorators(
					decor.Name(fmt.Sprintf("(%d/%d):%s%s", taskNum+1, totalSymbols, b.Name, table)),
					timeDecorator,
				),
				mpb.AppendDecorators(
					decor.Percentage(),
				),
			)

			lastKlineTime := startTime
			for {
				var getLastKline common.Kline
				result = b.DB.Table(table).Order("time DESC").First(&getLastKline)
				if result.Error != nil {
					if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
						global.Zap.Error(fmt.Sprintf("从数据库获取 %s 的最后时间失败:", table), zap.Error(result.Error))
						bar.Abort(true)
						return
					}
				} else {
					lastKlineTime = getLastKline.Time.Carbon
				}

				// 获取新的 K 线数据
				const klinesUrl = "/fapi/v1/klines"
				weight := 2
				b.checkLimitWeight(weight)

				var remoteKlines [][]interface{}
				resp, err := global.Resty.R().
					SetQueryParams(map[string]string{
						"symbol":    symbol,
						"interval":  interval,
						"startTime": strconv.FormatInt(lastKlineTime.TimestampMilli(), 10),
						"limit":     "499",
					}).
					SetResult(&remoteKlines).
					Get(b.BaseUrl + klinesUrl)

				if err != nil {
					global.Zap.Error(fmt.Sprintf("获取 %s 的K线数据失败:", symbol), zap.Error(err))
					bar.Abort(true)
					return
				}

				if resp.StatusCode() != 200 {
					global.Zap.Error(fmt.Sprintf("%s API 响应状态码: %d", b.Name, resp.StatusCode()), zap.String("response body", string(resp.Body())))
					bar.Abort(true)
					return
				}

				// 更新数据库
				var klines []common.Kline
				for _, remoteKline := range remoteKlines {
					kTime := carbon.DateTimeMilli{Carbon: global.Carbon.CreateFromTimestampMilli(int64(remoteKline[0].(float64)))}
					kOpen, _ := decimal.NewFromString(remoteKline[1].(string))
					kHigh, _ := decimal.NewFromString(remoteKline[2].(string))
					kLow, _ := decimal.NewFromString(remoteKline[3].(string))
					kClose, _ := decimal.NewFromString(remoteKline[4].(string))
					kVolume, _ := decimal.NewFromString(remoteKline[5].(string))
					kline := common.Kline{
						Time:   kTime,
						Open:   kOpen,
						High:   kHigh,
						Low:    kLow,
						Close:  kClose,
						Volume: kVolume,
					}
					klines = append(klines, kline)
				}
				err = b.DB.Table(table).Save(&klines).Error
				if err != nil {
					global.Zap.Error(fmt.Sprintf("更新 %s 的K线数据到数据库失败:", table), zap.Error(err))
					bar.Abort(true)
					return
				}

				remoteLastKlineTime := global.Carbon.CreateFromTimestampMilli(int64(remoteKlines[len(remoteKlines)-1][0].(float64)))

				// 更新进度条
				bar.SetCurrent((startTime.DiffInMinutes(remoteLastKlineTime) / utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes) * utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes)

				// 如果获取的 K 线数据时间已经接近现在，跳出循环
				if remoteLastKlineTime.DiffInMinutes(timeNow) < utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes {
					break
				}
			}
		}(job.Symbol, job.TaskNum, job.TotalSymbols, p)
	}
}

func (b *BinanceFuture) updateHistoryKLines(jobs <-chan Job, interval string, wg *sync.WaitGroup) {
	for job := range jobs {
		func(symbolInfo Symbol) {
			defer wg.Done()

			symbol := symbolInfo.Symbol
			table := fmt.Sprintf(`%s_%s`, global.Config.Mahakala.KlineInterval, symbol)

			// 获取数据库中最新的一条 K 线数据的时间，如果没有数据，那么 startTime 为上市时间，否则为最新数据的时间
			startTime := global.Carbon.CreateFromTimestampMilli(symbolInfo.OnboardDate)
			var lastKline common.Kline
			result := b.DB.Table(table).Order("time DESC").First(&lastKline)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					global.Zap.Error(fmt.Sprintf("从数据库获取 %s 的最后时间失败:", table), zap.Error(result.Error))
					return
				}
			} else {
				startTime = lastKline.Time.Carbon
			}

			timeNow := global.Carbon.Now()
			// 如果最后一条 K 线数据的时间距离现在不足一个 K 线周期，跳过
			if startTime.DiffInMinutes(timeNow) < utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes {
				return
			}
			lastKlineTime := startTime
			for {
				var getLastKline common.Kline
				result = b.DB.Table(table).Order("time DESC").First(&getLastKline)
				if result.Error != nil {
					if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
						global.Zap.Error(fmt.Sprintf("从数据库获取 %s 的最后时间失败:", table), zap.Error(result.Error))
						return
					}
				} else {
					lastKlineTime = getLastKline.Time.Carbon
				}

				// 获取新的 K 线数据
				const klinesUrl = "/fapi/v1/klines"
				weight := 2
				b.checkLimitWeight(weight)

				var remoteKlines [][]interface{}
				resp, err := global.Resty.R().
					SetQueryParams(map[string]string{
						"symbol":    symbol,
						"interval":  interval,
						"startTime": strconv.FormatInt(lastKlineTime.TimestampMilli(), 10),
						"limit":     "499",
					}).
					SetResult(&remoteKlines).
					Get(b.BaseUrl + klinesUrl)

				if err != nil {
					global.Zap.Error(fmt.Sprintf("获取 %s 的K线数据失败:", symbol), zap.Error(err))
					return
				}

				if resp.StatusCode() != 200 {
					global.Zap.Error(fmt.Sprintf("%s API 响应状态码: %d", b.Name, resp.StatusCode()), zap.String("response body", string(resp.Body())))
					return
				}

				// 更新数据库
				var klines []common.Kline
				for _, remoteKline := range remoteKlines {
					kTime := carbon.DateTimeMilli{Carbon: global.Carbon.CreateFromTimestampMilli(int64(remoteKline[0].(float64)))}
					kOpen, _ := decimal.NewFromString(remoteKline[1].(string))
					kHigh, _ := decimal.NewFromString(remoteKline[2].(string))
					kLow, _ := decimal.NewFromString(remoteKline[3].(string))
					kClose, _ := decimal.NewFromString(remoteKline[4].(string))
					kVolume, _ := decimal.NewFromString(remoteKline[5].(string))
					kline := common.Kline{
						Time:   kTime,
						Open:   kOpen,
						High:   kHigh,
						Low:    kLow,
						Close:  kClose,
						Volume: kVolume,
					}
					klines = append(klines, kline)
				}
				err = b.DB.Table(table).Save(&klines).Error
				if err != nil {
					global.Zap.Error(fmt.Sprintf("更新 %s 的K线数据到数据库失败:", table), zap.Error(err))
					return
				}

				remoteLastKlineTime := global.Carbon.CreateFromTimestampMilli(int64(remoteKlines[len(remoteKlines)-1][0].(float64)))

				// 如果获取的 K 线数据时间已经接近现在，跳出循环
				if remoteLastKlineTime.DiffInMinutes(timeNow) < utils.MapInterval[global.Config.Mahakala.KlineInterval].Minutes {
					break
				}
			}
		}(job.Symbol)
	}
}

func (b *BinanceFuture) GetName() string {
	return b.Name
}

func (b *BinanceFuture) GetSymbols() []string {
	var symbols []string
	for _, symbol := range b.Symbols {
		symbols = append(symbols, symbol.Symbol)
	}
	sort.Strings(symbols)
	return symbols
}

func (b *BinanceFuture) CheckSymbol(symbol string) bool {
	return b.SymbolsSet.Contains(symbol)
}

func (b *BinanceFuture) GetKlines(symbol, interval string) (klines []response.Kline, err error) {
	table := fmt.Sprintf(`%s_%s`, global.Config.Mahakala.KlineInterval, symbol)
	err = b.DB.Table(table).Select(`time_bucket(?, time) as period,
FIRST(open, time) AS open,
MAX(high) AS high,
MIN(low) as low,
LAST(close, time) AS close,
SUM(volume) AS volume`, utils.MapInterval[interval].String).Group("period").Order("period DESC").Limit(global.Config.Mahakala.AnalyzeAmount).Find(&klines).Error
	if err != nil {
		return nil, err
	}
	utils.ReverseKline(klines)
	return klines, nil
}
