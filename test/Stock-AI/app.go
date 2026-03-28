package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"stock-ai/backend/data"
	"stock-ai/backend/models"
	"stock-ai/backend/plugin"
	"stock-ai/backend/prompt"

	"gorm.io/gorm/clause"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	logFileOnce sync.Once
	logFile     *os.File
)

// App struct
type App struct {
	ctx             context.Context
	stockAPI        *data.StockAPI
	fundAPI         *data.FundAPI
	aiClient        *data.AIClient
	futuresAPI      *data.FuturesAPI
	globalMarketAPI *data.GlobalMarketAPI
	cryptoForexAPI  *data.CryptoForexAPI
	sentimentAPI    *data.SentimentAPI
	pluginManager   *plugin.Manager
	promptManager   *prompt.Manager
	// 自选股票价格缓存（用于提醒检查）
	stockPriceCache     map[string]*models.StockPrice
	stockPriceCacheLock sync.RWMutex
	stockKLineCache     map[string]map[string][]models.KLineData
	stockReportCache    map[string][]models.ResearchReport
	stockNoticeCache    map[string][]models.StockNotice
	stockFinancialCache map[string]*data.FinancialData
	stockDataCacheLock  sync.RWMutex
	fundPriceCache      map[string]*models.FundPrice
	fundPriceCacheLock  sync.RWMutex
}

type klineFetchSpec struct {
	period string
	count  int
}

const (
	appVersion                = "1.0.0"
	appBuildTime              = "2025-11-15"
	updateManifestURL         = "https://gitee.com/he-jun0000/Stock-AI/raw/main/releases/version.json"
	disableTradeLevelFallback = true
)

const updateScriptTemplate = `@echo off
setlocal
set TARGET=%~1
set SOURCE=%~2
set BACKUP=%TARGET%.bak
set LOGDIR=%USERPROFILE%\.stock-ai\logs
set LOGFILE=%LOGDIR%\update.log

if not exist "%LOGDIR%" mkdir "%LOGDIR%"
call :log "Updater started. target=%TARGET%"
timeout /t 1 >nul

if exist "%TARGET%" (
:backup
copy /Y "%TARGET%" "%BACKUP%" >nul
if errorlevel 1 (
  call :log "Waiting for application to exit..."
  timeout /t 1 >nul
  goto backup
)
call :log "Backup created at %BACKUP%"
)

:install
copy /Y "%SOURCE%" "%TARGET%" >nul
if errorlevel 1 (
  call :log "Install failed, restoring backup and retrying..."
  if exist "%BACKUP%" (
    copy /Y "%BACKUP%" "%TARGET%" >nul
  )
  timeout /t 1 >nul
  goto install
)
call :log "Install succeeded, launching new version"

start "" "%TARGET%"
del "%SOURCE%" >nul 2>&1
del "%BACKUP%" >nul 2>&1
call :log "Cleanup finished"
del "%~f0"
exit /b 0

:log
echo [%date% %time%] %~1>>"%LOGFILE%"
goto :eof
`

func init() {
	setupAppLogger()
}

func setupAppLogger() {
	logFileOnce.Do(func() {
		logDir := filepath.Join(getUserDataDir(), "logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Printf("创建日志目录失败: %v", err)
			return
		}

		logPath := filepath.Join(logDir, "app.log")
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("打开日志文件失败: %v", err)
			return
		}
		logFile = f
	})

	if logFile == nil {
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if path := logFile.Name(); path != "" {
		log.Printf("日志输出接管，写入: %s", path)
	}
}

func shutdownAppLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func appendTraceLog(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("%s", message)

	logDir := filepath.Join(getUserDataDir(), "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}

	traceFile := filepath.Join(logDir, "trace.log")
	f, err := os.OpenFile(traceFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(f, "[%s] %s\n", timestamp, message)
}

func getUserDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(homeDir, ".stock-ai")
}

func (a *App) logUpdateEvent(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[Update] %s", msg)

	logDir := filepath.Join(getUserDataDir(), "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}

	logFile := filepath.Join(logDir, "update.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, msg))
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		stockAPI:            data.NewStockAPI(),
		fundAPI:             data.NewFundAPI(),
		futuresAPI:          data.NewFuturesAPI(),
		globalMarketAPI:     data.NewGlobalMarketAPI(),
		cryptoForexAPI:      data.NewCryptoForexAPI(),
		sentimentAPI:        data.NewSentimentAPI(),
		stockPriceCache:     make(map[string]*models.StockPrice),
		stockKLineCache:     make(map[string]map[string][]models.KLineData),
		stockReportCache:    make(map[string][]models.ResearchReport),
		stockNoticeCache:    make(map[string][]models.StockNotice),
		stockFinancialCache: make(map[string]*data.FinancialData),
		fundPriceCache:      make(map[string]*models.FundPrice),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	setupAppLogger()
	a.ctx = ctx

	// 初始化数据库
	if err := data.InitDB(); err != nil {
		log.Printf("初始化数据库失败: %v", err)
	}

	// 初始化插件管理器
	pluginsDir := getPluginsDir()
	a.pluginManager = plugin.NewManager(pluginsDir)
	if err := a.pluginManager.Init(); err != nil {
		log.Printf("初始化插件管理器失败: %v", err)
	}

	// 初始化提示词管理器
	promptsDir := getPromptsDir()
	promptMgr, err := prompt.NewManager(promptsDir)
	if err != nil {
		log.Printf("初始化提示词管理器失败: %v", err)
	} else {
		a.promptManager = promptMgr
	}

	go a.prefetchWatchlistData()
	a.startPriceCacheUpdater()
}

// getPluginsDir 获取插件目录
func getPluginsDir() string {
	// 获取用户数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./plugins"
	}
	return filepath.Join(homeDir, ".stock-ai", "plugins")
}

// getPromptsDir 获取提示词目录
func getPromptsDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./prompts"
	}
	return filepath.Join(homeDir, ".stock-ai", "prompts")
}

// ========== 股票相关 ==========

// GetStockList 获取自选股票列表
func (a *App) GetStockList() ([]models.Stock, error) {
	var stocks []models.Stock
	err := data.GetDB().Find(&stocks).Error
	return stocks, err
}

// AddStock 添加自选股票
func (a *App) AddStock(code string) error {
	// 标准化股票代码
	code = normalizeStockCode(code)

	// 先检查是否存在（包括软删除的记录）
	var existingStock models.Stock
	err := data.GetDB().Unscoped().Where("code = ?", code).First(&existingStock).Error
	if err == nil {
		// 记录存在
		if existingStock.DeletedAt.Valid {
			// 是软删除的记录，恢复它
			if err := data.GetDB().Unscoped().Model(&existingStock).Updates(map[string]interface{}{
				"deleted_at": nil,
			}).Error; err != nil {
				return err
			}
			go a.prefetchStockData(code, nil)
			return nil
		}
		// 记录已存在且未删除
		return fmt.Errorf("股票 %s 已存在", code)
	}

	// 获取股票名称
	prices, err := a.stockAPI.GetStockPrice([]string{code})
	appendTraceLog("[TradeLevel] price fetch %s err=%v", code, err)
	if err != nil {
		return err
	}

	name := code
	if p, ok := prices[code]; ok && p.Name != "" {
		name = p.Name
	}

	market := "sh"
	if strings.HasPrefix(code, "sz") {
		market = "sz"
	}

	stock := models.Stock{
		Code:   code,
		Name:   name,
		Market: market,
	}

	if err := data.GetDB().Create(&stock).Error; err != nil {
		return err
	}

	go a.prefetchStockData(stock.Code, nil)
	return nil
}

// normalizeStockCode 标准化股票代码，自动添加市场前缀
func normalizeStockCode(code string) string {
	code = strings.TrimSpace(strings.ToLower(code))

	// 如果已经有前缀，直接返回
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") {
		return code
	}

	// 去除可能的点号前缀（如 000001.SZ）
	if strings.Contains(code, ".") {
		parts := strings.Split(code, ".")
		if len(parts) == 2 {
			numCode := parts[0]
			market := strings.ToLower(parts[1])
			if market == "sz" || market == "sh" {
				return market + numCode
			}
		}
	}

	// 根据代码规则自动判断市场
	// 上海：6开头（主板）、688开头（科创板）
	// 深圳：0开头（主板）、3开头（创业板）、002开头（中小板）
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") {
			return "sh" + code
		}
		if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
			return "sz" + code
		}
	}

	// 无法识别，默认返回原代码
	return code
}

// RemoveStock 删除自选股票（硬删除）
func (a *App) RemoveStock(code string) error {
	norm := normalizeStockCode(code)
	err := data.GetDB().Unscoped().Where("code = ?", norm).Delete(&models.Stock{}).Error
	if err == nil {
		a.clearStockCaches(norm)
	}
	return err
}

// GetStockPrice 获取股票实时价格
func (a *App) GetStockPrice(codes []string) (map[string]*models.StockPrice, error) {
	prices, err := a.stockAPI.GetStockPrice(codes)
	if err != nil {
		return prices, err
	}

	// 更新价格缓存（用于提醒检查）
	a.stockPriceCacheLock.Lock()
	for code, price := range prices {
		if price != nil {
			a.stockPriceCache[code] = price
		}
	}
	a.stockPriceCacheLock.Unlock()

	return prices, nil
}

// GetKLineData 获取K线数据
func (a *App) GetKLineData(code string, period string, count int) ([]models.KLineData, error) {
	code = normalizeStockCode(code)
	if period == "" {
		period = "daily"
	}
	appendTraceLog("[KLineTrace] 调用 GetKLineData code=%s period=%s count=%d", code, period, count)
	if data, ok := a.getCachedKLines(code, period, count); ok {
		appendTraceLog("[KLineTrace] 命中缓存 %s %s -> %d 条", code, period, len(data))
		return data, nil
	}

	klines, err := a.stockAPI.GetKLineData(code, period, count)
	if err != nil {
		if cached, ok := a.getCachedKLines(code, period, count); ok {
			appendTraceLog("[KLineTrace] 拉取失败，使用缓存数据 %s %s: %v", code, period, err)
			return cached, nil
		}
		appendTraceLog("[KLineTrace] 拉取失败且无缓存 %s %s: %v", code, period, err)
		return nil, err
	}
	if len(klines) == 0 {
		if cached, ok := a.getCachedKLines(code, period, count); ok {
			appendTraceLog("[KLineTrace] 数据为空，使用缓存数据 %s %s", code, period)
			return cached, nil
		}
		appendTraceLog("[KLineTrace] 数据为空且无缓存 %s %s", code, period)
		return nil, fmt.Errorf("K线数据为空: %s", code)
	}

	a.setKLineCache(code, period, klines)
	appendTraceLog("[KLineTrace] 实时获取成功 %s %s -> %d 条", code, period, len(klines))
	return cloneKLines(klines, count), nil
}

// GetMinuteData 获取分时数据
func (a *App) GetMinuteData(code string) ([]models.MinuteData, error) {
	return a.stockAPI.GetMinuteData(code)
}

// ========== 基金相关 ==========

// GetFundList 获取自选基金列表
func (a *App) GetFundList() ([]models.Fund, error) {
	var funds []models.Fund
	err := data.GetDB().Find(&funds).Error
	return funds, err
}

// AddFund 添加自选基金
func (a *App) AddFund(code string) error {
	prices, err := a.fundAPI.GetFundPrice([]string{code})
	if err != nil {
		return err
	}

	name := code
	if p, ok := prices[code]; ok && p.Name != "" {
		name = p.Name
	}

	fundType := ""
	if detail, err := a.fundAPI.GetFundDetail(code); err == nil && detail != nil {
		if detail.Name != "" {
			name = detail.Name
		}
		fundType = detail.Type
	}

	fund := models.Fund{
		Code: code,
		Name: name,
		Type: fundType,
	}

	return data.GetDB().Create(&fund).Error
}

// RemoveFund 删除自选基金
func (a *App) RemoveFund(code string) error {
	err := data.GetDB().Where("code = ?", code).Delete(&models.Fund{}).Error
	if err == nil {
		a.fundPriceCacheLock.Lock()
		delete(a.fundPriceCache, code)
		a.fundPriceCacheLock.Unlock()
	}
	return err
}

// GetFundPrice 获取基金估值
func (a *App) GetFundPrice(codes []string) (map[string]*models.FundPrice, error) {
	prices, err := a.fundAPI.GetFundPrice(codes)
	if err != nil {
		return prices, err
	}

	a.fundPriceCacheLock.Lock()
	for code, price := range prices {
		if price != nil {
			a.fundPriceCache[code] = price
		}
	}
	a.fundPriceCacheLock.Unlock()
	return prices, nil
}

// GetFundOverview 获取基金详情
func (a *App) GetFundOverview(code string) (*models.FundOverview, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("无效的基金代码")
	}

	priceResult, err := a.GetFundPrice([]string{code})
	if err != nil {
		return nil, fmt.Errorf("获取基金估值失败: %v", err)
	}
	price := priceResult[code]

	var (
		detail                      *models.FundDetail
		history                     []models.FundPerformancePoint
		stockHoldings, bondHoldings []models.FundHolding
		notices                     []models.FundNotice
		detailErr                   error
		historyErr                  error
		holdingErr                  error
		noticeErr                   error
	)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		detail, detailErr = a.fundAPI.GetFundDetail(code)
	}()

	go func() {
		defer wg.Done()
		history, historyErr = a.fundAPI.GetFundHistory(code, 120)
	}()

	go func() {
		defer wg.Done()
		stockHoldings, bondHoldings, holdingErr = a.fundAPI.GetFundHoldings(code)
	}()

	go func() {
		defer wg.Done()
		notices, noticeErr = a.fundAPI.GetFundNotices(code, 20)
	}()

	wg.Wait()

	if detailErr != nil || detail == nil {
		log.Printf("[Fund] 获取基金信息失败: %v", detailErr)
		if fallback := a.fundAPI.BuildFallbackFundDetail(code, price); fallback != nil {
			detail = fallback
		} else {
			return nil, fmt.Errorf("获取基金信息失败: %v", detailErr)
		}
	}

	if historyErr != nil {
		log.Printf("[Fund] 获取历史净值失败: %v", historyErr)
	}
	if holdingErr != nil {
		log.Printf("[Fund] 获取持仓失败: %v", holdingErr)
	}
	if noticeErr != nil {
		log.Printf("[Fund] 获取公告失败: %v", noticeErr)
	}

	return &models.FundOverview{
		Price:         price,
		Detail:        detail,
		History:       history,
		StockHoldings: stockHoldings,
		BondHoldings:  bondHoldings,
		Notices:       notices,
	}, nil
}

// ========== 期货相关 (暂时禁用，返回空数据) ==========

// GetFuturesProducts 获取期货品种列表
func (a *App) GetFuturesProducts() []models.FuturesProduct {
	// 暂时禁用，返回空数据
	return []models.FuturesProduct{}
}

// GetFuturesPrice 获取期货实时行情
func (a *App) GetFuturesPrice(codes []string) (map[string]*models.FuturesPrice, error) {
	// 暂时禁用，返回空数据
	return make(map[string]*models.FuturesPrice), nil
}

// GetMainContracts 获取主力合约列表
func (a *App) GetMainContracts() ([]models.FuturesPrice, error) {
	// 暂时禁用，返回空数据
	return []models.FuturesPrice{}, nil
}

// SearchFutures 搜索期货合约
func (a *App) SearchFutures(keyword string) ([]models.Futures, error) {
	// 暂时禁用，返回空数据
	return []models.Futures{}, nil
}

// GetFuturesList 获取自选期货列表
func (a *App) GetFuturesList() ([]models.Futures, error) {
	var futures []models.Futures
	err := data.GetDB().Find(&futures).Error
	return futures, err
}

// AddFutures 添加自选期货
func (a *App) AddFutures(code string, name string, exchange string) error {
	futures := models.Futures{
		Code:     code,
		Name:     name,
		Exchange: exchange,
		Product:  code,
	}
	return data.GetDB().Create(&futures).Error
}

// RemoveFutures 删除自选期货
func (a *App) RemoveFutures(code string) error {
	return data.GetDB().Where("code = ?", code).Delete(&models.Futures{}).Error
}

// ========== 美股相关 ==========

// GetPopularUSStocks 获取热门美股列表
func (a *App) GetPopularUSStocks() []models.USStock {
	return a.globalMarketAPI.GetPopularUSStocks()
}

// GetUSStockPrice 获取美股实时行情
func (a *App) GetUSStockPrice(symbols []string) (map[string]*models.USStockPrice, error) {
	return a.globalMarketAPI.GetUSStockPrice(symbols)
}

// SearchUSStock 搜索美股
func (a *App) SearchUSStock(keyword string) ([]models.USStock, error) {
	return a.globalMarketAPI.SearchUSStock(keyword)
}

// GetUSStockList 获取自选美股列表
func (a *App) GetUSStockList() ([]models.USStock, error) {
	var stocks []models.USStock
	err := data.GetDB().Find(&stocks).Error
	return stocks, err
}

// AddUSStock 添加自选美股
func (a *App) AddUSStock(symbol string, name string, nameCN string, exchange string) error {
	stock := models.USStock{
		Symbol:   symbol,
		Name:     name,
		NameCN:   nameCN,
		Exchange: exchange,
	}
	return data.GetDB().Create(&stock).Error
}

// RemoveUSStock 删除自选美股
func (a *App) RemoveUSStock(symbol string) error {
	return data.GetDB().Where("symbol = ?", symbol).Delete(&models.USStock{}).Error
}

// ========== 港股相关 ==========

// GetPopularHKStocks 获取热门港股列表
func (a *App) GetPopularHKStocks() []models.HKStock {
	return a.globalMarketAPI.GetPopularHKStocks()
}

// GetHKStockPrice 获取港股实时行情
func (a *App) GetHKStockPrice(codes []string) (map[string]*models.HKStockPrice, error) {
	return a.globalMarketAPI.GetHKStockPrice(codes)
}

// SearchHKStock 搜索港股
func (a *App) SearchHKStock(keyword string) ([]models.HKStock, error) {
	return a.globalMarketAPI.SearchHKStock(keyword)
}

// GetHKStockList 获取自选港股列表
func (a *App) GetHKStockList() ([]models.HKStock, error) {
	var stocks []models.HKStock
	err := data.GetDB().Find(&stocks).Error
	return stocks, err
}

// AddHKStock 添加自选港股
func (a *App) AddHKStock(code string, name string, nameCN string) error {
	stock := models.HKStock{
		Code:   code,
		Name:   name,
		NameCN: nameCN,
	}
	return data.GetDB().Create(&stock).Error
}

// RemoveHKStock 删除自选港股
func (a *App) RemoveHKStock(code string) error {
	return data.GetDB().Where("code = ?", code).Delete(&models.HKStock{}).Error
}

// ========== 全球指数相关 ==========

// GetGlobalIndicesList 获取全球指数列表
func (a *App) GetGlobalIndicesList() []models.GlobalIndex {
	return a.globalMarketAPI.GetGlobalIndicesList()
}

// GetGlobalIndices 获取全球指数实时行情
func (a *App) GetGlobalIndices() ([]models.GlobalIndex, error) {
	result, err := a.globalMarketAPI.GetGlobalIndices()
	if err == nil && len(result) > 0 {
		// 异步保存到持久化缓存
		go data.GetPersistentCache().SaveGlobalIndicesList(result)
	}
	return result, err
}

// GetGlobalNews 获取国际财经新闻
func (a *App) GetGlobalNews(country string) ([]models.NewsItem, error) {
	result, err := a.globalMarketAPI.GetGlobalNews(country)
	if err == nil && len(result) > 0 {
		go data.GetPersistentCache().SaveGlobalNews(country, result)
	}
	return result, err
}

// ========== 外汇相关 (暂时禁用，返回空数据) ==========

// GetMainForexPairs 获取主要外汇货币对列表
func (a *App) GetMainForexPairs() []models.ForexRate {
	// 暂时禁用，返回空数据
	return []models.ForexRate{}
}

// GetForexRates 获取外汇汇率
func (a *App) GetForexRates() ([]models.ForexRate, error) {
	// 暂时禁用，返回空数据
	return []models.ForexRate{}, nil
}

// ========== 市场情绪 ==========

// GetAShareSentiment 获取A股市场情绪
func (a *App) GetAShareSentiment() (*data.MarketSentiment, error) {
	return a.sentimentAPI.GetAShareSentiment()
}

// GetGlobalMarketSentiment 获取全球市场情绪
func (a *App) GetGlobalMarketSentiment(country string) (*data.MarketSentiment, error) {
	result, err := a.sentimentAPI.GetGlobalMarketSentiment(country)
	if err == nil && result != nil {
		go data.GetPersistentCache().SaveGlobalSentiment(country, result)
	}
	return result, err
}

// ========== 市场数据 ==========

// CachedMarketData 缓存的市场数据（用于前端）
type CachedMarketData struct {
	MarketIndex   []models.MarketIndex   `json:"marketIndex"`
	IndustryRank  []models.IndustryRank  `json:"industryRank"`
	MoneyFlow     []models.MoneyFlow     `json:"moneyFlow"`
	NewsList      []models.NewsItem      `json:"newsList"`
	LongTigerRank []models.LongTigerItem `json:"longTigerRank"`
	HotTopics     []models.HotTopic      `json:"hotTopics"`
	CacheTime     string                 `json:"cacheTime"`
	HasCache      bool                   `json:"hasCache"`
}

// CachedGlobalMarketData 缓存的全球市场数据（用于前端）
type CachedGlobalMarketData struct {
	GlobalIndices []models.GlobalIndex  `json:"globalIndices"`
	News          []models.NewsItem     `json:"news"`
	Sentiment     *data.MarketSentiment `json:"sentiment"`
	CacheTime     string                `json:"cacheTime"`
	HasCache      bool                  `json:"hasCache"`
}

// GetCachedMarketData 获取缓存的市场数据（启动时快速加载）
func (a *App) GetCachedMarketData() *CachedMarketData {
	pc := data.GetPersistentCache()
	cached, err := pc.LoadCache()
	if err != nil || cached == nil {
		return &CachedMarketData{HasCache: false}
	}

	return &CachedMarketData{
		MarketIndex:   cached.MarketIndex,
		IndustryRank:  cached.IndustryRank,
		MoneyFlow:     cached.MoneyFlow,
		NewsList:      cached.NewsList,
		LongTigerRank: cached.LongTigerRank,
		HotTopics:     cached.HotTopics,
		CacheTime:     cached.CacheTime.Format("2006-01-02 15:04:05"),
		HasCache:      true,
	}
}

// GetCachedGlobalMarketData 获取缓存的全球市场数据（启动时快速加载）
func (a *App) GetCachedGlobalMarketData(country string) *CachedGlobalMarketData {
	pc := data.GetPersistentCache()
	cached, err := pc.LoadCache()
	if err != nil || cached == nil {
		return &CachedGlobalMarketData{HasCache: false}
	}

	result := &CachedGlobalMarketData{
		GlobalIndices: cached.GlobalIndicesList,
		CacheTime:     cached.CacheTime.Format("2006-01-02 15:04:05"),
		HasCache:      len(cached.GlobalIndicesList) > 0,
	}

	// 获取该国家的新闻
	if cached.GlobalNews != nil {
		if news, ok := cached.GlobalNews[country]; ok {
			result.News = news
		}
	}

	// 获取该国家的情绪
	if cached.GlobalSentiment != nil {
		if sentiment, ok := cached.GlobalSentiment[country]; ok {
			result.Sentiment = sentiment
		}
	}

	return result
}

// MarkFirstLoadComplete 标记首次加载完成，切换到轮询模式
func (a *App) MarkFirstLoadComplete() {
	data.GetMultiSourceManager().SetFirstLoadComplete()
	log.Println("[App] 首次加载完成，切换到轮询模式")
}

// IsFirstLoad 检查是否为首次加载
func (a *App) IsFirstLoad() bool {
	return data.GetMultiSourceManager().IsFirstLoad()
}

// GetMarketIndex 获取市场指数
func (a *App) GetMarketIndex() ([]models.MarketIndex, error) {
	result, err := a.stockAPI.GetMarketIndex()
	if err == nil && len(result) > 0 {
		// 异步保存到持久化缓存
		go data.GetPersistentCache().SaveMarketIndex(result)
	}
	return result, err
}

// GetIndustryRank 获取行业排行
func (a *App) GetIndustryRank() ([]models.IndustryRank, error) {
	result, err := a.stockAPI.GetIndustryRank()
	if err == nil && len(result) > 0 {
		go data.GetPersistentCache().SaveIndustryRank(result)
	}
	return result, err
}

// GetMoneyFlow 获取资金流向
func (a *App) GetMoneyFlow() ([]models.MoneyFlow, error) {
	result, err := a.stockAPI.GetMoneyFlow()
	if err == nil && len(result) > 0 {
		go data.GetPersistentCache().SaveMoneyFlow(result)
	}
	return result, err
}

// GetNewsList 获取新闻快讯
func (a *App) GetNewsList() ([]models.NewsItem, error) {
	result, err := a.stockAPI.GetNewsList()
	if err == nil && len(result) > 0 {
		go data.GetPersistentCache().SaveNewsList(result)
	}
	return result, err
}

// GetResearchReports 获取研报列表
func (a *App) GetResearchReports(stockCode string) ([]models.ResearchReport, error) {
	code := normalizeStockCode(stockCode)
	return a.getResearchReportsCached(code)
}

// GetStockNotices 获取公告列表
func (a *App) GetStockNotices(stockCode string) ([]models.StockNotice, error) {
	code := normalizeStockCode(stockCode)
	return a.getStockNoticesCached(code)
}

// GetLongTigerRank 获取龙虎榜
func (a *App) GetLongTigerRank() ([]models.LongTigerItem, error) {
	result, err := a.stockAPI.GetLongTigerRank()
	if err == nil && len(result) > 0 {
		go data.GetPersistentCache().SaveLongTigerRank(result)
	}
	return result, err
}

// GetHotTopics 获取热门话题
func (a *App) GetHotTopics() ([]models.HotTopic, error) {
	result, err := a.stockAPI.GetHotTopics()
	if err == nil && len(result) > 0 {
		go data.GetPersistentCache().SaveHotTopics(result)
	}
	return result, err
}

// ========== 配置相关 ==========

// GetConfig 获取配置
func (a *App) GetConfig() (*models.Config, error) {
	var config models.Config
	err := data.GetDB().First(&config).Error
	if err == nil {
		// 更新请求管理器的配置
		data.GetRequestManager().UpdateConfig(&config)
	}
	return &config, err
}

// SaveConfig 保存配置
func (a *App) SaveConfig(config models.Config) error {
	err := data.GetDB().Save(&config).Error
	if err == nil {
		// 更新请求管理器的配置
		data.GetRequestManager().UpdateConfig(&config)
	}
	return err
}

// SkipUpdateVersion 记录跳过的版本
func (a *App) SkipUpdateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("版本号不能为空")
	}

	config, err := a.GetConfig()
	if err != nil {
		return err
	}
	if config.SkipUpdateVersion == version {
		return nil
	}

	config.SkipUpdateVersion = version
	return a.SaveConfig(*config)
}

// clearSkipUpdateVersion 清除跳过的版本记录
func (a *App) clearSkipUpdateVersion() error {
	config, err := a.GetConfig()
	if err != nil {
		return err
	}
	if config.SkipUpdateVersion == "" {
		return nil
	}
	config.SkipUpdateVersion = ""
	return a.SaveConfig(*config)
}

// ========== 交易时间相关 ==========

// TradingTimeInfo 交易时间信息
type TradingTimeInfo struct {
	IsTradingTime   bool `json:"isTradingTime"`
	IsPreMarketTime bool `json:"isPreMarketTime"`
	RefreshInterval int  `json:"refreshInterval"`
}

// GetTradingTimeInfo 获取交易时间信息
func (a *App) GetTradingTimeInfo() *TradingTimeInfo {
	// 获取配置的刷新间隔
	var config models.Config
	baseInterval := 15
	if err := data.GetDB().First(&config).Error; err == nil {
		if config.RefreshInterval > 0 {
			baseInterval = config.RefreshInterval
		}
	}

	return &TradingTimeInfo{
		IsTradingTime:   data.IsTradingTime(),
		IsPreMarketTime: data.IsPreMarketTime(),
		RefreshInterval: data.GetRefreshInterval(baseInterval),
	}
}

// GetDataPipelineStatus 获取数据通道总览
func (a *App) GetDataPipelineStatus() (*models.DataPipelineStatus, error) {
	msm := data.GetMultiSourceManager()
	pipeline := &models.DataPipelineStatus{
		MarketSources: msm.GetStatusList(),
		Financial:     data.GetFinancialClient().GetDataSourceStatus(),
		Proxy:         data.GetRequestManager().GetProxyStatus(),
		GeneratedAt:   time.Now().Format(time.RFC3339),
	}
	return pipeline, nil
}

// FrontendTrace 允许前端写入调试日志
func (a *App) FrontendTrace(message string) {
	if strings.TrimSpace(message) == "" {
		return
	}
	appendTraceLog("[Frontend] %s", message)
}

func (a *App) startPriceCacheUpdater() {
	go func() {
		for {
			if err := a.refreshPriceCache(); err != nil {
				log.Printf("[PriceCache] 刷新失败: %v", err)
			}
			time.Sleep(a.getPriceCacheInterval())
		}
	}()
}

func (a *App) prefetchWatchlistData() {
	stocks, err := a.GetStockList()
	if err != nil || len(stocks) == 0 {
		return
	}

	log.Printf("[Prefetch] 开始预加载 %d 条自选股数据", len(stocks))

	var cfg *models.Config
	if config, err := a.GetConfig(); err == nil {
		cfg = config
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, stock := range stocks {
		code := normalizeStockCode(stock.Code)
		if code == "" {
			continue
		}

		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			a.prefetchStockData(code, cfg)
		}(code)
	}

	wg.Wait()
	log.Printf("[Prefetch] 自选股数据预加载完成")
}

func (a *App) prefetchStockData(stockCode string, cfg *models.Config) {
	code := normalizeStockCode(stockCode)
	if code == "" {
		return
	}

	if prices, err := a.stockAPI.GetStockPrice([]string{code}); err == nil {
		if price, ok := prices[code]; ok && price != nil {
			a.stockPriceCacheLock.Lock()
			a.stockPriceCache[code] = price
			a.stockPriceCacheLock.Unlock()
		}
	}

	if klines, err := a.stockAPI.GetKLineData(code, "daily", 120); err == nil && len(klines) > 0 {
		a.setKLineCache(code, "daily", klines)
	}

	if reports, err := a.stockAPI.GetResearchReports(code); err == nil {
		a.stockDataCacheLock.Lock()
		a.stockReportCache[code] = reports
		a.stockDataCacheLock.Unlock()
	}

	if notices, err := a.stockAPI.GetStockNotices(code); err == nil {
		a.stockDataCacheLock.Lock()
		a.stockNoticeCache[code] = notices
		a.stockDataCacheLock.Unlock()
	}

	activeCfg := cfg
	if activeCfg == nil {
		if config, err := a.GetConfig(); err == nil {
			activeCfg = config
		}
	}
	if activeCfg != nil {
		if fin, err := a.fetchFinancialData(code, activeCfg); err == nil && fin != nil {
			a.stockDataCacheLock.Lock()
			a.stockFinancialCache[code] = fin
			a.stockDataCacheLock.Unlock()
		}
	}
}

type stockAnalysisData struct {
	stock      *models.StockPrice
	klines     []models.KLineData
	reports    []models.ResearchReport
	notices    []models.StockNotice
	position   *models.Position
	financial  *data.FinancialData
	warmupDone bool
}

type warmupStep struct {
	name     string
	duration time.Duration
	err      error
}

func (a *App) warmupStockAnalysisData(code, analysisType string, cfg *models.Config, timeout time.Duration) *stockAnalysisData {
	result := &stockAnalysisData{}
	var wg sync.WaitGroup

	var steps []*warmupStep
	var stepsMu sync.Mutex

	addTask := func(name string, enabled bool, fn func() error) {
		if !enabled {
			return
		}

		step := &warmupStep{name: name}
		stepsMu.Lock()
		steps = append(steps, step)
		stepsMu.Unlock()

		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			if err := fn(); err != nil {
				step.err = err
			}
			step.duration = time.Since(start)
			log.Printf("[Prefetch] %s 完成，耗时 %s，err=%v", name, step.duration, step.err)
		}()
	}

	addTask("price", true, func() error {
		stock, err := a.getPriceSnapshot(code)
		if err == nil {
			result.stock = stock
		}
		return err
	})

	addTask("kline", true, func() error {
		klines, err := a.getKLineDataCached(code, 120)
		if err == nil {
			result.klines = klines
		}
		return err
	})

	addTask("reports", true, func() error {
		reports, err := a.getResearchReportsCached(code)
		if err == nil {
			result.reports = reports
		}
		return err
	})

	addTask("notices", true, func() error {
		notices, err := a.getStockNoticesCached(code)
		if err == nil {
			result.notices = notices
		}
		return err
	})

	addTask("position", true, func() error {
		position, err := a.GetPositionByStock(code)
		if err == nil {
			result.position = position
		}
		return err
	})

	needFinancial := analysisType == "fundamental" || analysisType == "master"
	addTask("financial", needFinancial, func() error {
		fin, err := a.getFinancialDataCached(code, cfg)
		if err == nil {
			result.financial = fin
		}
		return err
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	start := time.Now()
	if timeout <= 0 {
		<-done
		result.warmupDone = true
		log.Printf("[Prefetch] 股票 %s 预热完成，总耗时 %s", code, time.Since(start))
		return result
	}

	select {
	case <-done:
		result.warmupDone = true
		log.Printf("[Prefetch] 股票 %s 预热完成，总耗时 %s", code, time.Since(start))
	case <-time.After(timeout):
		log.Printf("[Prefetch] 股票 %s 预热超过 %v，终止", code, timeout)
	}

	return result
}

func (a *App) getPriceCacheInterval() time.Duration {
	interval := 15
	if config, err := a.GetConfig(); err == nil && config.RefreshInterval > 0 {
		interval = config.RefreshInterval
	}
	if interval <= 0 {
		interval = 15
	}
	return time.Duration(interval) * time.Second
}

func (a *App) refreshPriceCache() error {
	codes := make(map[string]struct{})
	fundCodes := make(map[string]struct{})

	var stocks []models.Stock
	if err := data.GetDB().Find(&stocks).Error; err == nil {
		for _, s := range stocks {
			if s.Code != "" {
				codes[s.Code] = struct{}{}
			}
		}
	} else {
		log.Printf("[PriceCache] 加载自选股失败: %v", err)
	}

	var alerts []models.StockAlert
	if err := data.GetDB().Find(&alerts).Error; err == nil {
		for _, alert := range alerts {
			if alert.StockCode != "" {
				codes[alert.StockCode] = struct{}{}
			}
		}
	} else {
		log.Printf("[PriceCache] 加载提醒失败: %v", err)
	}

	if len(codes) > 0 {
		codeList := make([]string, 0, len(codes))
		for code := range codes {
			codeList = append(codeList, code)
		}

		prices, err := a.stockAPI.GetStockPrice(codeList)
		if err != nil {
			return err
		}

		a.stockPriceCacheLock.Lock()
		for code, price := range prices {
			if price != nil {
				a.stockPriceCache[code] = price
			}
		}
		a.stockPriceCacheLock.Unlock()
	}

	var funds []models.Fund
	if err := data.GetDB().Find(&funds).Error; err == nil {
		for _, f := range funds {
			if f.Code != "" {
				fundCodes[f.Code] = struct{}{}
			}
		}
	} else {
		log.Printf("[PriceCache] 加载自选基金失败: %v", err)
	}

	var fundAlerts []models.FundAlert
	if err := data.GetDB().Find(&fundAlerts).Error; err == nil {
		for _, alert := range fundAlerts {
			if alert.FundCode != "" {
				fundCodes[alert.FundCode] = struct{}{}
			}
		}
	} else {
		log.Printf("[PriceCache] 加载基金提醒失败: %v", err)
	}

	if len(fundCodes) > 0 {
		codeList := make([]string, 0, len(fundCodes))
		for code := range fundCodes {
			codeList = append(codeList, code)
		}

		prices, err := a.fundAPI.GetFundPrice(codeList)
		if err != nil {
			log.Printf("[PriceCache] 获取基金估值失败: %v", err)
		} else {
			a.fundPriceCacheLock.Lock()
			for code, price := range prices {
				if price != nil {
					a.fundPriceCache[code] = price
				}
			}
			a.fundPriceCacheLock.Unlock()
		}
	}

	return nil
}

func isAllowedWebContentURL(target string) bool {
	parsed, err := url.Parse(target)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	host := parsed.Hostname()
	if host == "" {
		return false
	}
	host = strings.ToLower(host)
	if host == "localhost" {
		return false
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
			return false
		}
	}
	return true
}

func (a *App) getPriceSnapshot(code string) (*models.StockPrice, error) {
	a.stockPriceCacheLock.RLock()
	if price, ok := a.stockPriceCache[code]; ok && price != nil {
		a.stockPriceCacheLock.RUnlock()
		return price, nil
	}
	a.stockPriceCacheLock.RUnlock()

	prices, err := a.stockAPI.GetStockPrice([]string{code})
	if err != nil {
		return nil, err
	}

	price, ok := prices[code]
	if !ok || price == nil {
		return nil, fmt.Errorf("未找到股票: %s", code)
	}

	a.stockPriceCacheLock.Lock()
	a.stockPriceCache[code] = price
	a.stockPriceCacheLock.Unlock()
	return price, nil
}

func (a *App) getKLineDataCached(code string, count int) ([]models.KLineData, error) {
	return a.getKLineDataCachedByPeriod(code, "daily", count)
}

func (a *App) getKLineDataCachedByPeriod(code, period string, count int) ([]models.KLineData, error) {
	if klines, ok := a.getCachedKLines(code, period, count); ok {
		log.Printf("[KLine] getKLineDataCached 命中缓存 %s %s -> %d 条", code, period, len(klines))
		return klines, nil
	}

	limit := count
	if min := getMinKLineFetchCount(period); limit < min {
		limit = min
	}

	klines, err := a.stockAPI.GetKLineData(code, period, limit)
	if err != nil {
		if cached, ok := a.getCachedKLines(code, period, count); ok {
			return cached, nil
		}
		return nil, err
	}
	if len(klines) > 0 {
		a.setKLineCache(code, period, klines)
		return cloneKLines(klines, count), nil
	}
	if cached, ok := a.getCachedKLines(code, period, count); ok {
		log.Printf("[KLine] getKLineDataCached 使用回退缓存 %s %s -> %d 条", code, period, len(cached))
		return cached, nil
	}
	return nil, fmt.Errorf("K线数据为空: %s [%s]", code, period)
}

func getMinKLineFetchCount(period string) int {
	switch period {
	case "week":
		return 60
	case "month":
		return 36
	default:
		return 60
	}
}

func (a *App) getCachedKLines(code string, period string, count int) ([]models.KLineData, bool) {
	a.stockDataCacheLock.RLock()
	defer a.stockDataCacheLock.RUnlock()
	if cache, ok := a.stockKLineCache[code]; ok {
		if klines, ok := cache[period]; ok && len(klines) > 0 {
			return cloneKLines(klines, count), true
		}
	}
	return nil, false
}

func (a *App) setKLineCache(code string, period string, klines []models.KLineData) {
	if len(klines) == 0 {
		return
	}
	copied := make([]models.KLineData, len(klines))
	copy(copied, klines)
	a.stockDataCacheLock.Lock()
	if _, ok := a.stockKLineCache[code]; !ok {
		a.stockKLineCache[code] = make(map[string][]models.KLineData)
	}
	a.stockKLineCache[code][period] = copied
	a.stockDataCacheLock.Unlock()
}

func (a *App) getResearchReportsCached(code string) ([]models.ResearchReport, error) {
	a.stockDataCacheLock.RLock()
	if reports, ok := a.stockReportCache[code]; ok && len(reports) > 0 {
		copied := make([]models.ResearchReport, len(reports))
		copy(copied, reports)
		a.stockDataCacheLock.RUnlock()
		return copied, nil
	}
	a.stockDataCacheLock.RUnlock()

	reports, err := a.stockAPI.GetResearchReports(code)
	if err != nil {
		a.stockDataCacheLock.RLock()
		if cached, ok := a.stockReportCache[code]; ok && len(cached) > 0 {
			copied := make([]models.ResearchReport, len(cached))
			copy(copied, cached)
			a.stockDataCacheLock.RUnlock()
			return copied, nil
		}
		a.stockDataCacheLock.RUnlock()
		return nil, err
	}
	if len(reports) == 0 {
		a.stockDataCacheLock.RLock()
		if cached, ok := a.stockReportCache[code]; ok && len(cached) > 0 {
			copied := make([]models.ResearchReport, len(cached))
			copy(copied, cached)
			a.stockDataCacheLock.RUnlock()
			return copied, nil
		}
		a.stockDataCacheLock.RUnlock()
		return reports, nil
	}
	a.stockDataCacheLock.Lock()
	a.stockReportCache[code] = reports
	a.stockDataCacheLock.Unlock()
	return reports, nil
}

func (a *App) getStockNoticesCached(code string) ([]models.StockNotice, error) {
	a.stockDataCacheLock.RLock()
	if notices, ok := a.stockNoticeCache[code]; ok && len(notices) > 0 {
		copied := make([]models.StockNotice, len(notices))
		copy(copied, notices)
		a.stockDataCacheLock.RUnlock()
		return copied, nil
	}
	a.stockDataCacheLock.RUnlock()

	notices, err := a.stockAPI.GetStockNotices(code)
	if err != nil {
		a.stockDataCacheLock.RLock()
		if cached, ok := a.stockNoticeCache[code]; ok && len(cached) > 0 {
			copied := make([]models.StockNotice, len(cached))
			copy(copied, cached)
			a.stockDataCacheLock.RUnlock()
			return copied, nil
		}
		a.stockDataCacheLock.RUnlock()
		return nil, err
	}
	if len(notices) == 0 {
		a.stockDataCacheLock.RLock()
		if cached, ok := a.stockNoticeCache[code]; ok && len(cached) > 0 {
			copied := make([]models.StockNotice, len(cached))
			copy(copied, cached)
			a.stockDataCacheLock.RUnlock()
			return copied, nil
		}
		a.stockDataCacheLock.RUnlock()
		return notices, nil
	}
	a.stockDataCacheLock.Lock()
	a.stockNoticeCache[code] = notices
	a.stockDataCacheLock.Unlock()
	return notices, nil
}

func (a *App) getFinancialDataCached(code string, cfg *models.Config) (*data.FinancialData, error) {
	a.stockDataCacheLock.RLock()
	if fin, ok := a.stockFinancialCache[code]; ok && fin != nil {
		a.stockDataCacheLock.RUnlock()
		return fin, nil
	}
	a.stockDataCacheLock.RUnlock()

	activeCfg := cfg
	if activeCfg == nil {
		if conf, err := a.GetConfig(); err == nil {
			activeCfg = conf
		} else {
			return nil, err
		}
	}

	fin, err := a.fetchFinancialData(code, activeCfg)
	if err != nil {
		return nil, err
	}
	if fin != nil {
		a.stockDataCacheLock.Lock()
		a.stockFinancialCache[code] = fin
		a.stockDataCacheLock.Unlock()
	}
	return fin, nil
}

func cloneKLines(src []models.KLineData, count int) []models.KLineData {
	if len(src) == 0 {
		return nil
	}
	if count <= 0 || count > len(src) {
		count = len(src)
	}
	start := len(src) - count
	if start < 0 {
		start = 0
	}
	result := make([]models.KLineData, count)
	copy(result, src[start:])
	return result
}

const tradeLevelPriceTimeout = 5 * time.Second

func (a *App) lookupStockName(code string) string {
	var stock models.Stock
	if err := data.GetDB().Where("code = ?", code).First(&stock).Error; err == nil {
		if strings.TrimSpace(stock.Name) != "" {
			return stock.Name
		}
	}

	a.stockPriceCacheLock.RLock()
	if cached, ok := a.stockPriceCache[code]; ok && cached != nil && strings.TrimSpace(cached.Name) != "" {
		name := cached.Name
		a.stockPriceCacheLock.RUnlock()
		return name
	}
	a.stockPriceCacheLock.RUnlock()
	return strings.ToUpper(code)
}

func (a *App) buildPriceFromKLines(code string, daily []models.KLineData) *models.StockPrice {
	if len(daily) == 0 {
		return nil
	}
	last := daily[len(daily)-1]
	preClose := last.Close
	if len(daily) > 1 {
		preClose = daily[len(daily)-2].Close
	}
	change := last.Close - preClose
	changePercent := 0.0
	if preClose > 0 {
		changePercent = (change / preClose) * 100
	}

	return &models.StockPrice{
		Code:          code,
		Name:          a.lookupStockName(code),
		Price:         last.Close,
		Open:          last.Open,
		High:          last.High,
		Low:           last.Low,
		PreClose:      preClose,
		Change:        change,
		ChangePercent: changePercent,
		Volume:        last.Volume,
		Amount:        0,
		UpdateTime:    last.Date,
	}
}

func (a *App) resolveTradeLevelPrice(code string, daily []models.KLineData) (*models.StockPrice, error) {
	prices, err := a.fetchStockPriceWithTimeout(code, tradeLevelPriceTimeout)
	if err != nil {
		appendTraceLog("[TradeLevel] price realtime err %s: %v", code, err)
	} else {
		appendTraceLog("[TradeLevel] price realtime fetch %s -> %d", code, len(prices))
	}
	if price := prices[code]; price != nil {
		appendTraceLog("[TradeLevel] price realtime hit %s", code)
		return price, nil
	}
	if err == nil && len(prices) == 0 {
		appendTraceLog("[TradeLevel] price realtime empty %s", code)
	} else if err == nil {
		appendTraceLog("[TradeLevel] price realtime miss %s", code)
	}

	if disableTradeLevelFallback {
		return nil, fmt.Errorf("实时行情不可用，请稍后再试: %v", err)
	}

	a.stockPriceCacheLock.RLock()
	if cached, ok := a.stockPriceCache[code]; ok && cached != nil {
		a.stockPriceCacheLock.RUnlock()
		appendTraceLog("[TradeLevel] price fallback cache %s", code)
		return cached, nil
	}
	a.stockPriceCacheLock.RUnlock()

	if fallback := a.buildPriceFromKLines(code, daily); fallback != nil {
		appendTraceLog("[TradeLevel] price fallback kline %s", code)
		return fallback, nil
	}
	return nil, fmt.Errorf("无法获取股票实时价格，请稍后再试")
}

func (a *App) fetchStockPriceWithTimeout(code string, timeout time.Duration) (map[string]*models.StockPrice, error) {
	type priceResult struct {
		price  *models.StockPrice
		source string
		err    error
	}

	sources := []struct {
		name  string
		fetch func([]string) (map[string]*models.StockPrice, error)
	}{
		{"Primary", a.stockAPI.GetStockPrice},
		{"Tencent", a.stockAPI.GetStockPriceFromTencent},
		{"Eastmoney", a.stockAPI.GetStockPriceFromEastmoney},
		{"Netease", a.stockAPI.GetStockPriceFromNetease},
		{"Sohu", a.stockAPI.GetStockPriceFromSohu},
		{"Xueqiu", a.stockAPI.GetStockPriceFromXueqiu},
		{"Baidu", a.stockAPI.GetStockPriceFromBaidu},
		{"Hexun", a.stockAPI.GetStockPriceFromHexun},
	}

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	resultCh := make(chan priceResult, len(sources))

	var wg sync.WaitGroup
	for _, src := range sources {
		src := src
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					select {
					case <-ctx.Done():
					case resultCh <- priceResult{err: fmt.Errorf("panic: %v", r), source: src.name}:
					}
				}
			}()
			prices, err := src.fetch([]string{code})
			res := priceResult{source: src.name}
			if err != nil {
				res.err = err
			} else if prices != nil {
				if price := prices[code]; price != nil {
					res.price = price
				} else {
					res.err = fmt.Errorf("%s 返回空数据", src.name)
				}
			} else {
				res.err = fmt.Errorf("%s 未返回数据", src.name)
			}
			select {
			case <-ctx.Done():
				return
			case resultCh <- res:
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var lastErr error
	for {
		select {
		case <-ctx.Done():
			if lastErr == nil {
				lastErr = fmt.Errorf("所有数据源均超时或无结果")
			}
			return nil, lastErr
		case res, ok := <-resultCh:
			if !ok {
				if lastErr == nil {
					lastErr = fmt.Errorf("所有数据源均无结果")
				}
				return nil, lastErr
			}
			if res.price != nil {
				appendTraceLog("[TradeLevel] price realtime via %s", res.source)
				cancel()
				return map[string]*models.StockPrice{code: res.price}, nil
			}
			if res.err != nil {
				lastErr = res.err
				appendTraceLog("[TradeLevel] price source err %s %s: %v", code, res.source, res.err)
			}
		}
	}
}

func logTradeLevelFail(code, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	appendTraceLog("[TradeLevel] fail %s: %s", code, msg)
}

func (a *App) PrefetchTradeLevelData(code string) {
	code = normalizeStockCode(code)
	appendTraceLog("[TradeLevel] Prefetch request %s", code)
	if code == "" {
		return
	}
	go func(stockCode string) {
		log.Printf("[Prefetch] ?? %s ?????K???", stockCode)
		specs := []klineFetchSpec{
			{period: "daily", count: 240},
			{period: "week", count: 180},
			{period: "month", count: 180},
		}
		for _, spec := range specs {
			if _, err := a.getKLineDataCachedByPeriod(stockCode, spec.period, spec.count); err != nil {
				log.Printf("[Prefetch] %s %s ????: %v", stockCode, spec.period, err)
				appendTraceLog("[TradeLevel] Prefetch fail %s %s: %v", stockCode, spec.period, err)
			} else {
				appendTraceLog("[TradeLevel] Prefetch hit %s %s", stockCode, spec.period)
			}
		}
		log.Printf("[Prefetch] ?? %s ???K???", stockCode)
		appendTraceLog("[TradeLevel] Prefetch done %s", stockCode)
	}(code)
}

func (a *App) clearStockCaches(code string) {
	a.stockPriceCacheLock.Lock()
	delete(a.stockPriceCache, code)
	a.stockPriceCacheLock.Unlock()

	a.stockDataCacheLock.Lock()
	delete(a.stockKLineCache, code)
	delete(a.stockReportCache, code)
	delete(a.stockNoticeCache, code)
	delete(a.stockFinancialCache, code)
	a.stockDataCacheLock.Unlock()
}

func (a *App) fetchFinancialData(code string, cfg *models.Config) (*data.FinancialData, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不可用")
	}
	financialClient := data.GetFinancialClient()
	if cfg.TushareToken != "" {
		financialClient.SetTushareToken(cfg.TushareToken)
	}
	financialClient.SetPreferTushare(cfg.DataSourcePriority != "akshare")
	return financialClient.GetFinancialData(code)
}

// ClearCache 清除缓存
func (a *App) ClearCache() {
	data.GetRequestManager().ClearAllCache()
}

func (a *App) getProAnalysisCache(code, analysisType, masterStyle string) *models.ProAnalysisCache {
	if code == "" {
		return nil
	}
	code = normalizeStockCode(code)

	var cache models.ProAnalysisCache
	query := data.GetDB().Where("stock_code = ? AND analysis_type = ? AND master_style = ?", code, analysisType, masterStyle)
	if err := query.Order("updated_at DESC").First(&cache).Error; err != nil {
		return nil
	}
	return &cache
}

func (a *App) saveProAnalysisCache(code, analysisType, masterStyle, content string) {
	code = normalizeStockCode(code)
	if code == "" {
		return
	}
	text := strings.TrimSpace(content)
	if text == "" {
		return
	}

	cache := models.ProAnalysisCache{
		StockCode:    code,
		AnalysisType: analysisType,
		MasterStyle:  masterStyle,
		Content:      text,
		UpdatedAt:    time.Now(),
	}

	if err := data.GetDB().Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "stock_code"},
			{Name: "analysis_type"},
			{Name: "master_style"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"content":    cache.Content,
			"updated_at": cache.UpdatedAt,
		}),
	}).Create(&cache).Error; err != nil {
		log.Printf("[AI分析] 缓存写入失败: %v", err)
	}
}

// GetProAnalysisCache 返回指定股票/类型的大师分析缓存
func (a *App) GetProAnalysisCache(code, analysisType, masterStyle string) *models.ProAnalysisCache {
	code = normalizeStockCode(code)
	if code == "" {
		return nil
	}
	if cache := a.getProAnalysisCache(code, analysisType, masterStyle); cache != nil {
		return cache
	}
	if masterStyle != "" {
		return a.getProAnalysisCache(code, analysisType, "")
	}
	return nil
}

// ========== 系统相关 ==========

// GetVersion 获取版本信息
func (a *App) GetVersion() *models.VersionInfo {
	return &models.VersionInfo{
		Version:   appVersion,
		BuildTime: appBuildTime,
	}
}

// CheckUpdate 检查更新
func (a *App) CheckUpdate() *models.UpdateInfo {
	currentVersion := appVersion

	var skipVersion string
	if config, err := a.GetConfig(); err == nil {
		skipVersion = config.SkipUpdateVersion
	}

	// 从远程获取最新版本信息
	// 这里使用一个简单的 JSON 文件作为版本检测源
	// 你可以将这个文件托管在 GitHub Pages 或其他静态服务器上
	updateUrl := updateManifestURL

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(updateUrl)
	if err != nil {
		// 网络错误，返回无更新
		return &models.UpdateInfo{
			HasUpdate:  false,
			CurrentVer: currentVersion,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &models.UpdateInfo{
			HasUpdate:  false,
			CurrentVer: currentVersion,
		}
	}

	var versionData struct {
		Version     string `json:"version"`
		Description string `json:"description"`
		DownloadUrl string `json:"downloadUrl"`
		ReleaseUrl  string `json:"releaseUrl"`
		ReleaseDate string `json:"releaseDate"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&versionData); err != nil {
		return &models.UpdateInfo{
			HasUpdate:  false,
			CurrentVer: currentVersion,
		}
	}

	// 比较版本号
	hasUpdate := compareVersions(versionData.Version, currentVersion) > 0
	skipped := false

	if skipVersion != "" {
		cmp := compareVersions(versionData.Version, skipVersion)
		if cmp == 0 {
			skipped = true
			hasUpdate = false
		} else if cmp > 0 {
			if err := a.clearSkipUpdateVersion(); err != nil {
				log.Printf("清除跳过版本失败: %v", err)
			}
			skipVersion = ""
		}
	}

	return &models.UpdateInfo{
		HasUpdate:   hasUpdate,
		Version:     versionData.Version,
		CurrentVer:  currentVersion,
		Description: versionData.Description,
		DownloadUrl: versionData.DownloadUrl,
		ReleaseUrl:  versionData.ReleaseUrl,
		ReleaseDate: versionData.ReleaseDate,
		SkipVersion: skipVersion,
		Skipped:     skipped,
	}
}

// DownloadAndInstallUpdate 自动下载并安装更新
func (a *App) DownloadAndInstallUpdate() (*models.UpdateInfo, error) {
	info := a.CheckUpdate()
	if info == nil {
		return nil, fmt.Errorf("无法获取更新信息")
	}

	if !info.HasUpdate || info.DownloadUrl == "" {
		return info, fmt.Errorf("当前已是最新版本")
	}

	a.logUpdateEvent("Preparing to update from %s to %s", info.CurrentVer, info.Version)
	go a.handleAutoUpdate(info)
	return info, nil
}

func (a *App) handleAutoUpdate(info *models.UpdateInfo) {
	if err := a.downloadAndInstall(info); err != nil {
		a.logUpdateEvent("Update failed: %v", err)
		a.emitUpdateStatus("error", err.Error())
	}
}

func (a *App) downloadAndInstall(info *models.UpdateInfo) error {
	a.emitUpdateStatus("downloading", "正在下载更新包...")
	a.logUpdateEvent("Start downloading update package: %s", info.Version)

	if err := a.clearSkipUpdateVersion(); err != nil {
		log.Printf("清除跳过版本失败: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "stock-ai-update")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	downloadURL, err := url.Parse(info.DownloadUrl)
	if err != nil {
		return fmt.Errorf("无效的下载地址: %w", err)
	}

	fileName := path.Base(downloadURL.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = fmt.Sprintf("stock-ai-%s.exe", info.Version)
	}

	tempFilePath := filepath.Join(tmpDir, fileName)
	out, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	client := &http.Client{
		Timeout: 0,
	}
	resp, err := client.Get(info.DownloadUrl)
	if err != nil {
		return fmt.Errorf("下载更新失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载更新失败，状态码: %d", resp.StatusCode)
	}

	const megaByte = 1024.0 * 1024.0
	total := resp.ContentLength
	var downloaded int64
	startTime := time.Now()
	buf := make([]byte, 64*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("写入临时文件失败: %w", writeErr)
			}
			downloaded += int64(n)

			var percent float64
			if total > 0 {
				percent = (float64(downloaded) / float64(total)) * 100
			}

			var msg string
			if total > 0 {
				msg = fmt.Sprintf("已下载 %.2f / %.2f MB", float64(downloaded)/megaByte, float64(total)/megaByte)
			} else {
				msg = fmt.Sprintf("已下载 %.2f MB", float64(downloaded)/megaByte)
			}
			elapsed := time.Since(startTime).Seconds()
			speedMBps := 0.0
			etaSeconds := -1.0
			if elapsed > 0 {
				speedMBps = (float64(downloaded) / megaByte) / elapsed
			}
			if total > 0 && speedMBps > 0 {
				remainingMB := (float64(total-downloaded) / megaByte)
				etaSeconds = remainingMB / speedMBps
			}

			a.emitUpdateProgress(percent, msg, speedMBps, etaSeconds)
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("下载更新失败: %w", readErr)
		}
	}

	a.emitUpdateProgress(100, "下载完成", 0, 0)
	a.emitUpdateStatus("installing", "下载完成，正在安装...")
	a.logUpdateEvent("Update package downloaded successfully: %s", tempFilePath)

	if err := out.Close(); err != nil {
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取程序路径: %w", err)
	}

	scriptPath := filepath.Join(tmpDir, "apply_update.bat")
	if err := os.WriteFile(scriptPath, []byte(updateScriptTemplate), 0644); err != nil {
		return fmt.Errorf("创建更新脚本失败: %w", err)
	}

	cmd := exec.Command("cmd", "/C", "call", scriptPath, exePath, tempFilePath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("执行更新脚本失败: %w", err)
	}

	a.emitUpdateStatus("restarting", "更新完成，应用即将重新启动...")
	a.logUpdateEvent("Update script launched, restarting application")
	go func() {
		time.Sleep(2 * time.Second)
		wailsRuntime.Quit(a.ctx)
	}()

	return nil
}

func (a *App) emitUpdateProgress(percent float64, message string, speed float64, etaSeconds float64) {
	payload := map[string]interface{}{
		"percent": percent,
		"message": message,
	}

	if speed > 0 {
		payload["speed"] = speed
	}
	if etaSeconds >= 0 {
		payload["etaSeconds"] = etaSeconds
	}

	wailsRuntime.EventsEmit(a.ctx, "update:progress", payload)
}

func (a *App) emitUpdateStatus(status string, message string) {
	a.logUpdateEvent("%s - %s", strings.ToUpper(status), message)
	wailsRuntime.EventsEmit(a.ctx, "update:status", map[string]interface{}{
		"status":  status,
		"message": message,
	})
}

// compareVersions 比较版本号，返回 1 表示 v1 > v2，-1 表示 v1 < v2，0 表示相等
func compareVersions(v1, v2 string) int {
	// 移除 v 前缀
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}

// OpenURL 打开URL
func (a *App) OpenURL(url string) {
	wailsRuntime.BrowserOpenURL(a.ctx, url)
}

// ========== AI 相关 ==========

// AIChat AI对话（非流式）
func (a *App) AIChat(request models.AIChatRequest) (*models.AIChatResponse, error) {
	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		return nil, fmt.Errorf("获取配置失败: %v", err)
	}

	if !config.AiEnabled {
		return nil, fmt.Errorf("AI功能未启用，请在设置中开启")
	}

	if config.AiApiKey == "" {
		return nil, fmt.Errorf("请先配置AI API Key")
	}

	// 创建或更新AI客户端
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	// 构建消息
	messages := []data.ChatMessage{
		{Role: "system", Content: data.BuildChatSystemPrompt()},
		{Role: "user", Content: request.Message},
	}

	// 如果指定了股票或基金代码，添加上下文
	if request.StockCode != "" {
		stockContext, err := a.buildStockContext(request.StockCode)
		if err == nil && stockContext != "" {
			messages = []data.ChatMessage{
				{Role: "system", Content: data.BuildChatSystemPrompt()},
				{Role: "user", Content: stockContext + "\n\n用户问题：" + request.Message},
			}
		}
	} else if request.FundCode != "" {
		fundContext, err := a.buildFundContext(request.FundCode)
		if err == nil && fundContext != "" {
			messages = []data.ChatMessage{
				{Role: "system", Content: data.BuildChatSystemPrompt()},
				{Role: "user", Content: fundContext + "\n\n用户问题：" + request.Message},
			}
		}
	}

	// 调用AI
	content, err := a.aiClient.ChatWithTimeout(messages, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("AI调用失败: %v", err)
	}

	return &models.AIChatResponse{
		Content:   content,
		SessionID: request.SessionID,
		Done:      true,
	}, nil
}

// AIChatStream AI对话（流式）- 通过事件推送
func (a *App) AIChatStream(request models.AIChatRequest) error {
	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "获取配置失败")
		return fmt.Errorf("获取配置失败: %v", err)
	}

	if !config.AiEnabled {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "AI功能未启用，请在设置中开启")
		return fmt.Errorf("AI功能未启用")
	}

	if config.AiApiKey == "" {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "请先配置AI API Key")
		return fmt.Errorf("请先配置AI API Key")
	}

	// 创建或更新AI客户端
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	// 生成会话ID（如果没有提供）
	sessionID := request.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("chat_%d", time.Now().UnixNano())
	}

	// 保存用户消息到数据库
	userMsg := models.AIMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   request.Message,
	}
	data.GetDB().Create(&userMsg)

	// 构建消息
	messages := []data.ChatMessage{
		{Role: "system", Content: data.BuildChatSystemPrompt()},
		{Role: "user", Content: request.Message},
	}

	// 如果指定了股票或基金代码，添加上下文
	if request.StockCode != "" {
		stockContext, err := a.buildStockContext(request.StockCode)
		if err == nil && stockContext != "" {
			messages = []data.ChatMessage{
				{Role: "system", Content: data.BuildChatSystemPrompt()},
				{Role: "user", Content: stockContext + "\n\n用户问题：" + request.Message},
			}
		}
	} else if request.FundCode != "" {
		fundContext, err := a.buildFundContext(request.FundCode)
		if err == nil && fundContext != "" {
			messages = []data.ChatMessage{
				{Role: "system", Content: data.BuildChatSystemPrompt()},
				{Role: "user", Content: fundContext + "\n\n用户问题：" + request.Message},
			}
		}
	}

	// 启动流式调用
	go func() {
		ch, err := a.aiClient.ChatStream(messages)
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", err.Error())
			return
		}

		var fullResponse strings.Builder
		for content := range ch {
			fullResponse.WriteString(content)
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-stream", content)
		}

		// 保存AI回复到数据库
		aiMsg := models.AIMessage{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fullResponse.String(),
		}
		data.GetDB().Create(&aiMsg)

		wailsRuntime.EventsEmit(a.ctx, "ai-chat-done", "")
	}()

	return nil
}

// AIAnalyzeStock AI分析股票
func (a *App) AIAnalyzeStock(code string) (*models.AIChatResponse, error) {
	// 标准化股票代码
	code = normalizeStockCode(code)

	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		return nil, fmt.Errorf("获取配置失败: %v", err)
	}

	if !config.AiEnabled {
		return nil, fmt.Errorf("AI功能未启用，请在设置中开启")
	}

	if config.AiApiKey == "" {
		return nil, fmt.Errorf("请先配置AI API Key")
	}

	// 创建AI客户端
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	// 获取股票数据
	prices, err := a.stockAPI.GetStockPrice([]string{code})
	if err != nil {
		return nil, fmt.Errorf("获取股票价格失败: %v", err)
	}

	stock, ok := prices[code]
	if !ok {
		return nil, fmt.Errorf("未找到股票 %s", code)
	}

	// 获取K线数据
	klines, _ := a.stockAPI.GetKLineData(code, "daily", 30)

	// 获取研报
	reports, _ := a.stockAPI.GetResearchReports(code)

	// 获取公告
	notices, _ := a.stockAPI.GetStockNotices(code)

	// 构建分析提示词
	prompt := data.BuildStockAnalysisPrompt(stock, klines, reports, notices)

	// 调用AI
	messages := []data.ChatMessage{
		{Role: "system", Content: data.BuildChatSystemPrompt()},
		{Role: "user", Content: prompt},
	}

	content, err := a.aiClient.ChatWithTimeout(messages, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %v", err)
	}

	return &models.AIChatResponse{
		Content: content,
		Done:    true,
	}, nil
}

// AIAnalyzeStockStream AI分析股票（流式）
func (a *App) AIAnalyzeStockStream(code string) error {
	// 标准化股票代码
	code = normalizeStockCode(code)

	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "获取配置失败")
		return err
	}

	if !config.AiEnabled {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "AI功能未启用，请在设置中开启")
		return fmt.Errorf("AI功能未启用")
	}

	if config.AiApiKey == "" {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "请先配置AI API Key")
		return fmt.Errorf("请先配置AI API Key")
	}

	// 创建AI客户端
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	go func() {
		// 获取股票数据
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-stream", "正在获取股票数据...\n\n")

		prices, err := a.stockAPI.GetStockPrice([]string{code})
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "获取股票价格失败")
			return
		}

		stock, ok := prices[code]
		if !ok {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "未找到股票")
			return
		}

		// 获取K线数据
		klines, _ := a.stockAPI.GetKLineData(code, "daily", 30)

		// 获取研报
		reports, _ := a.stockAPI.GetResearchReports(code)

		// 获取公告
		notices, _ := a.stockAPI.GetStockNotices(code)

		// 构建分析提示词
		prompt := data.BuildStockAnalysisPrompt(stock, klines, reports, notices)

		// 调用AI
		messages := []data.ChatMessage{
			{Role: "system", Content: data.BuildChatSystemPrompt()},
			{Role: "user", Content: prompt},
		}

		ch, err := a.aiClient.ChatStream(messages)
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", err.Error())
			return
		}

		for content := range ch {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-stream", content)
		}
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-done", "")
	}()

	return nil
}

// AIAnalyzeTradeLevels 计算AI买卖区间
func (a *App) AIAnalyzeTradeLevels(code string) (*models.TradeLevelResult, error) {
	code = normalizeStockCode(code)
	appendTraceLog("[TradeLevel] start %s", code)

	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		logTradeLevelFail(code, "获取配置失败: %v", err)
		return nil, fmt.Errorf("获取配置失败: %v", err)
	}
	if !config.AiEnabled {
		logTradeLevelFail(code, "AI功能未启用")
		return nil, fmt.Errorf("AI功能未启用，请在设置中开启")
	}
	if config.AiApiKey == "" {
		logTradeLevelFail(code, "缺少AI API Key")
		return nil, fmt.Errorf("请先配置AI API Key")
	}
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	specs := []klineFetchSpec{
		{period: "daily", count: 180},
		{period: "week", count: 180},
		{period: "month", count: 180},
	}

	results := make(map[string][]models.KLineData, len(specs))
	appendTraceLog("[TradeLevel] fetching K? %s", code)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, spec := range specs {
		wg.Add(1)
		go func(s klineFetchSpec) {
			defer wg.Done()
			klines, err := a.getKLineDataCachedByPeriod(code, s.period, s.count)
			if err != nil {
				log.Printf("[AI] 获取 %s K线失败: %v", s.period, err)
				return
			}
			mu.Lock()
			results[s.period] = klines
			mu.Unlock()
		}(spec)
	}
	wg.Wait()
	appendTraceLog("[TradeLevel] K??? %s ?=%d ?=%d ?=%d", code, len(results["daily"]), len(results["week"]), len(results["month"]))

	daily := results["daily"]
	weekly := results["week"]
	monthly := results["month"]
	if len(daily) == 0 {
		logTradeLevelFail(code, "日线数据为空，无法生成AI提示")
		return nil, fmt.Errorf("暂时无法获取该股票的K线数据，请稍后重试")
	}

	stock, err := a.resolveTradeLevelPrice(code, daily)
	if err != nil {
		logTradeLevelFail(code, "获取价格失败: %v", err)
		return nil, err
	}

	prompt := data.BuildTradeLevelPrompt(stock, daily, weekly, monthly)
	appendTraceLog("[TradeLevel] prompt ready %s", code)
	messages := []data.ChatMessage{
		{Role: "system", Content: data.BuildChatSystemPrompt()},
		{Role: "user", Content: prompt},
	}

	appendTraceLog("[TradeLevel] call AI %s", code)
	content, err := a.aiClient.ChatWithTimeout(messages, 30*time.Second)
	if err != nil {
		logTradeLevelFail(code, "AI调用失败: %v", err)
		return nil, fmt.Errorf("AI计算失败: %v", err)
	}

	result, err := data.ParseTradeLevelResponse(content)
	if err != nil {
		appendTraceLog("[TradeLevel] parse fail %s: %v", code, err)
		logTradeLevelFail(code, "AI结果解析失败: %v", err)
		return nil, err
	}
	appendTraceLog("[TradeLevel] success %s", code)
	return result, nil
}

// AIAnalyzeFundStream AI分析基金（流式）
func (a *App) AIAnalyzeFundStream(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("请输入基金代码")
	}

	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "获取配置失败")
		return err
	}
	if !config.AiEnabled {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "AI功能未启用，请在设置中开启")
		return fmt.Errorf("AI功能未启用")
	}
	if config.AiApiKey == "" {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "请先配置AI API Key")
		return fmt.Errorf("请先配置AI API Key")
	}
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	go func() {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-stream", "正在获取基金数据...\n\n")
		overview, err := a.GetFundOverview(code)
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", fmt.Sprintf("获取基金数据失败: %v", err))
			return
		}

		prompt := data.BuildFundAnalysisPrompt(
			overview.Detail,
			overview.Price,
			overview.History,
			overview.StockHoldings,
			overview.Notices,
		)

		messages := []data.ChatMessage{
			{Role: "system", Content: data.BuildChatSystemPrompt()},
			{Role: "user", Content: prompt},
		}

		ch, err := a.aiClient.ChatStream(messages)
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", err.Error())
			return
		}

		for content := range ch {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-stream", content)
		}
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-done", "")
	}()

	return nil
}

// AIRecommend AI选股推荐
func (a *App) AIRecommend() (*models.AIChatResponse, error) {
	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		return nil, fmt.Errorf("获取配置失败: %v", err)
	}

	if !config.AiEnabled {
		return nil, fmt.Errorf("AI功能未启用，请在设置中开启")
	}

	if config.AiApiKey == "" {
		return nil, fmt.Errorf("请先配置AI API Key")
	}

	// 创建AI客户端
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	// 获取市场数据
	indexes, _ := a.stockAPI.GetMarketIndex()
	industries, _ := a.stockAPI.GetIndustryRank()
	moneyFlow, _ := a.stockAPI.GetMoneyFlow()

	// 构建推荐提示词
	prompt := data.BuildRecommendPrompt(indexes, industries, moneyFlow)

	// 调用AI
	messages := []data.ChatMessage{
		{Role: "system", Content: data.BuildChatSystemPrompt()},
		{Role: "user", Content: prompt},
	}

	content, err := a.aiClient.Chat(messages)
	if err != nil {
		return nil, fmt.Errorf("AI推荐失败: %v", err)
	}

	return &models.AIChatResponse{
		Content: content,
		Done:    true,
	}, nil
}

// AIRecommendStream AI选股推荐（流式）
func (a *App) AIRecommendStream() error {
	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "获取配置失败")
		return err
	}

	if !config.AiEnabled {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "AI功能未启用，请在设置中开启")
		return fmt.Errorf("AI功能未启用")
	}

	if config.AiApiKey == "" {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", "请先配置AI API Key")
		return fmt.Errorf("请先配置AI API Key")
	}

	// 创建AI客户端
	if a.aiClient == nil {
		a.aiClient = data.NewAIClient(&config)
	}

	// 生成会话ID
	sessionID := fmt.Sprintf("market_%d", time.Now().UnixNano())

	// 保存用户消息到数据库
	userMsg := models.AIMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   "根据当前市场数据给出选股推荐",
	}
	data.GetDB().Create(&userMsg)

	go func() {
		wailsRuntime.EventsEmit(a.ctx, "ai-chat-stream", "正在分析市场数据...\n\n")

		// 优先从本地缓存获取市场数据，避免重新请求
		var indexes []models.MarketIndex
		var industries []models.IndustryRank
		var moneyFlow []models.MoneyFlow

		pc := data.GetPersistentCache()
		cached, err := pc.LoadCache()
		if err == nil && cached != nil && len(cached.MarketIndex) > 0 {
			// 使用缓存数据
			indexes = cached.MarketIndex
			industries = cached.IndustryRank
			moneyFlow = cached.MoneyFlow
			log.Printf("[AI推荐] 使用本地缓存数据，缓存时间: %v", cached.CacheTime)
		} else {
			// 缓存不可用，重新获取
			log.Printf("[AI推荐] 缓存不可用，重新获取数据")
			indexes, _ = a.stockAPI.GetMarketIndex()
			industries, _ = a.stockAPI.GetIndustryRank()
			moneyFlow, _ = a.stockAPI.GetMoneyFlow()
		}

		// 构建推荐提示词
		prompt := data.BuildRecommendPrompt(indexes, industries, moneyFlow)

		// 调用AI
		messages := []data.ChatMessage{
			{Role: "system", Content: data.BuildChatSystemPrompt()},
			{Role: "user", Content: prompt},
		}

		ch, err := a.aiClient.ChatStream(messages)
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-error", err.Error())
			return
		}

		var fullResponse strings.Builder
		fullResponse.WriteString("正在分析市场数据...\n\n")
		for content := range ch {
			fullResponse.WriteString(content)
			wailsRuntime.EventsEmit(a.ctx, "ai-chat-stream", content)
		}

		// 保存AI回复到数据库
		aiMsg := models.AIMessage{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fullResponse.String(),
		}
		data.GetDB().Create(&aiMsg)

		wailsRuntime.EventsEmit(a.ctx, "ai-chat-done", "")
	}()

	return nil
}

// AISummarizeContent AI摘要内容（流式）
// 参数: title, contentType, pageURL, infoCode(研报), artCode(公告), stockCode, manualContent
func (a *App) AISummarizeContentStream(title string, contentType string, pageURL string, infoCode string, artCode string, stockCode string, manualContent string) error {
	log.Printf("[AI摘要] 开始处理: title=%s, type=%s, url=%s, infoCode=%s, artCode=%s", title, contentType, pageURL, infoCode, artCode)

	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		log.Printf("[AI摘要] 获取配置失败: %v", err)
		wailsRuntime.EventsEmit(a.ctx, "ai-summary-error", "获取配置失败")
		return err
	}

	log.Printf("[AI摘要] 配置: enabled=%v, model=%s, hasKey=%v", config.AiEnabled, config.AiModel, config.AiApiKey != "")

	if !config.AiEnabled {
		wailsRuntime.EventsEmit(a.ctx, "ai-summary-error", "AI功能未启用，请在设置中开启")
		return fmt.Errorf("AI功能未启用")
	}

	if config.AiApiKey == "" {
		wailsRuntime.EventsEmit(a.ctx, "ai-summary-error", "请先配置AI API Key")
		return fmt.Errorf("请先配置AI API Key")
	}

	// 每次都重新创建AI客户端，确保使用最新配置
	a.aiClient = data.NewAIClient(&config)
	log.Printf("[AI摘要] AI客户端已创建")

	// 获取手动输入的内容
	userContent := manualContent

	// 生成会话ID
	sessionID := fmt.Sprintf("summary_%d", time.Now().UnixNano())

	// 保存用户消息到数据库
	userMsgContent := fmt.Sprintf("请分析%s：%s", contentType, title)
	userMsg := models.AIMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   userMsgContent,
	}
	data.GetDB().Create(&userMsg)

	go func() {
		log.Printf("[AI摘要] goroutine开始执行")
		var webContent string
		var fetchMethod string

		// 优先使用手动输入的内容
		if userContent != "" {
			webContent = userContent
			fetchMethod = "手动输入"
			wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "使用手动输入的内容进行分析...\n\n")
		} else {
			// 阶梯式获取内容
			wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "正在获取内容...\n")

			// 方案1: 使用东方财富API获取内容
			if contentType == "report" && infoCode != "" {
				log.Printf("[AI摘要] 尝试API获取研报...")
				wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "尝试通过API获取研报内容...\n")
				content, err := a.stockAPI.GetReportContent(infoCode)
				if err == nil && content != "" {
					webContent = content
					fetchMethod = "东方财富API"
					wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "成功获取研报内容！\n\n")
				} else {
					log.Printf("[AI摘要] API获取研报失败: %v", err)
				}
			} else if contentType == "notice" && artCode != "" {
				log.Printf("[AI摘要] 尝试API获取公告...")
				wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "尝试通过API获取公告内容...\n")
				content, err := a.stockAPI.GetNoticeContent(stockCode, artCode)
				if err == nil && content != "" {
					webContent = content
					fetchMethod = "东方财富API"
					wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "成功获取公告内容！\n\n")
				} else {
					log.Printf("[AI摘要] API获取公告失败: %v", err)
				}
			}

			// 方案2: 使用本地浏览器获取内容
			if webContent == "" && pageURL != "" {
				if !isAllowedWebContentURL(pageURL) {
					log.Printf("[AI摘要] 拒绝不安全的URL: %s", pageURL)
				} else {
					log.Printf("[AI摘要] 尝试浏览器获取内容: %s", pageURL)
					wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "尝试通过本地浏览器获取内容...\n")
					browserFetcher := data.NewBrowserFetcher(config.BrowserPath)
					if browserFetcher.IsAvailable() {
						log.Printf("[AI摘要] %s 可用，开始抓取...", browserFetcher.Name())
						content, err := browserFetcher.FetchContent(pageURL)
						log.Printf("[AI摘要] 浏览器抓取完成: len=%d, err=%v", len(content), err)
						if err == nil && content != "" && len(content) > 100 {
							webContent = content
							fetchMethod = browserFetcher.Name() + "抓取"
							wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "成功获取页面内容！\n\n")
						} else {
							log.Printf("[AI摘要] 浏览器获取失败: %v", err)
						}
					} else {
						log.Printf("[AI摘要] 未检测到可用浏览器")
					}
				}
			}

			// 如果都失败了，基于标题分析
			if webContent == "" {
				wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", "无法获取详细内容，将基于标题进行分析...\n\n")
				fetchMethod = "标题分析"
			}
		}

		// 构建摘要提示词
		var prompt string
		if webContent != "" && fetchMethod != "标题分析" {
			// 有实际内容，进行深度分析
			if contentType == "report" {
				prompt = fmt.Sprintf(`请对以下研报进行专业分析：

研报标题：%s

研报内容：
%s

请从以下几个方面进行深度分析：
1. **核心观点**：研报的主要结论和观点
2. **关键数据**：报告中提到的重要数据和指标
3. **投资建议**：研报给出的投资评级和目标价（如有）
4. **风险因素**：报告中提到的风险点
5. **总结**：用2-3句话概括研报要点

请用简洁专业的语言进行分析。`, title, webContent)
			} else {
				prompt = fmt.Sprintf(`请对以下公告进行专业分析：

公告标题：%s

公告内容：
%s

请从以下几个方面进行深度分析：
1. **公告类型**：这是什么类型的公告
2. **核心内容**：公告的主要事项和关键信息
3. **影响分析**：对公司经营和股价可能产生的影响
4. **关注要点**：投资者应该重点关注的内容
5. **总结**：用2-3句话概括公告要点

请用简洁专业的语言进行分析。`, title, webContent)
			}
		} else {
			// 没有内容，基于标题分析
			if contentType == "report" {
				prompt = fmt.Sprintf(`请根据以下研报标题，分析并给出该研报可能包含的关键信息：

研报标题：%s

请从以下几个方面进行分析：
1. **研报主题**：这份研报主要讨论什么内容
2. **关键观点**：根据标题推测研报可能的核心观点
3. **投资建议**：可能的投资评级和建议
4. **关注要点**：投资者应该关注的重点
5. **风险提示**：可能存在的风险因素

请用简洁专业的语言进行分析。`, title)
			} else {
				prompt = fmt.Sprintf(`请根据以下公告标题，分析并给出该公告可能包含的关键信息：

公告标题：%s

请从以下几个方面进行分析：
1. **公告类型**：这是什么类型的公告
2. **核心内容**：公告可能涉及的主要事项
3. **影响分析**：对公司和股价可能产生的影响
4. **关注要点**：投资者应该关注的重点
5. **后续跟踪**：需要持续关注的事项

请用简洁专业的语言进行分析。`, title)
			}
		}

		// 调用AI
		log.Printf("[AI摘要] 准备调用AI, prompt长度: %d", len(prompt))
		messages := []data.ChatMessage{
			{Role: "system", Content: "你是一位专业的证券分析师，擅长解读研报和公告。请进行专业、客观的分析。"},
			{Role: "user", Content: prompt},
		}

		log.Printf("[AI摘要] 开始调用ChatStream...")
		ch, err := a.aiClient.ChatStream(messages)
		if err != nil {
			log.Printf("[AI摘要] ChatStream失败: %v", err)
			wailsRuntime.EventsEmit(a.ctx, "ai-summary-error", err.Error())
			return
		}
		log.Printf("[AI摘要] ChatStream成功，开始接收响应...")

		var fullResponse strings.Builder
		contentCount := 0
		for content := range ch {
			contentCount++
			fullResponse.WriteString(content)
			wailsRuntime.EventsEmit(a.ctx, "ai-summary-stream", content)
		}
		log.Printf("[AI摘要] 接收完成，共收到 %d 个内容块", contentCount)

		// 保存AI回复到数据库
		aiMsg := models.AIMessage{
			SessionID: sessionID,
			Role:      "assistant",
			Content:   fullResponse.String(),
		}
		data.GetDB().Create(&aiMsg)

		wailsRuntime.EventsEmit(a.ctx, "ai-summary-done", "")
	}()

	return nil
}

// buildStockContext 构建股票上下文
func (a *App) buildStockContext(code string) (string, error) {
	code = normalizeStockCode(code)

	stock, err := a.getPriceSnapshot(code)
	if err != nil {
		return "", err
	}

	klines, _ := a.getKLineDataCached(code, 10)
	reports, _ := a.getResearchReportsCached(code)
	notices, _ := a.getStockNoticesCached(code)

	return data.BuildStockAnalysisPrompt(stock, klines, reports, notices), nil
}

func (a *App) buildFundContext(code string) (string, error) {
	overview, err := a.GetFundOverview(code)
	if err != nil {
		return "", err
	}
	return data.BuildFundAnalysisPrompt(
		overview.Detail,
		overview.Price,
		overview.History,
		overview.StockHoldings,
		overview.Notices,
	), nil
}

// AIAnalyzeByType 按类型分析股票（流式）
// analysisType: fundamental(基本面), technical(技术面), sentiment(情绪面), master(大师模式)
// masterStyle: buffett(巴菲特), lynch(彼得林奇), graham(格雷厄姆), liverta(利弗莫尔)
func (a *App) AIAnalyzeByTypeStream(code string, analysisType string, masterStyle string) error {
	code = normalizeStockCode(code)
	log.Printf("[专业分析] 开始: code=%s, type=%s, master=%s", code, analysisType, masterStyle)

	// 检查AI是否启用
	var config models.Config
	if err := data.GetDB().First(&config).Error; err != nil {
		wailsRuntime.EventsEmit(a.ctx, "ai-analysis-error", "获取配置失败")
		return err
	}

	if !config.AiEnabled {
		wailsRuntime.EventsEmit(a.ctx, "ai-analysis-error", "AI功能未启用，请在设置中开启")
		return fmt.Errorf("AI功能未启用")
	}

	if config.AiApiKey == "" {
		wailsRuntime.EventsEmit(a.ctx, "ai-analysis-error", "请先配置AI API Key")
		return fmt.Errorf("请先配置AI API Key")
	}

	a.aiClient = data.NewAIClient(&config)

	// 为确保分析基于最新行情，先强制刷新该股票的行情/财务缓存
	a.prefetchStockData(code, &config)

	preloadChan := make(chan *stockAnalysisData, 1)
	go func() {
		preloadChan <- a.warmupStockAnalysisData(code, analysisType, &config, 5*time.Second)
	}()

	go func(preloadResult <-chan *stockAnalysisData, stockCode, aType, style string) {
		wailsRuntime.EventsEmit(a.ctx, "ai-analysis-stream", "AI已开始准备数据，稍后将持续输出，请勿关闭窗口...\n\n")

		var preloaded *stockAnalysisData
		select {
		case preloaded = <-preloadResult:
		case <-time.After(300 * time.Millisecond):
			preloaded = &stockAnalysisData{}
			go func() {
				<-preloadResult
			}()
		}
		if preloaded == nil {
			preloaded = &stockAnalysisData{}
		}

		stock := preloaded.stock
		if stock == nil {
			var err error
			stock, err = a.getPriceSnapshot(stockCode)
			if err != nil {
				wailsRuntime.EventsEmit(a.ctx, "ai-analysis-error", "获取股票价格失败")
				return
			}
		}

		klines := preloaded.klines
		if len(klines) == 0 {
			var err error
			klines, err = a.getKLineDataCached(code, 60)
			if err != nil {
				log.Printf("[AI分析] 获取K线失败: %v", err)
			}
		} else if len(klines) > 60 {
			klines = cloneKLines(klines, 60)
		}

		reports := preloaded.reports
		if reports == nil {
			if r, err := a.getResearchReportsCached(code); err == nil {
				reports = r
			}
		}

		notices := preloaded.notices
		if notices == nil {
			if n, err := a.getStockNoticesCached(code); err == nil {
				notices = n
			}
		}

		position := preloaded.position
		if position == nil {
			position, _ = a.GetPositionByStock(code)
		}

		var financialData *data.FinancialData
		if aType == "fundamental" || aType == "master" {
			financialData = preloaded.financial
			if financialData == nil {
				wailsRuntime.EventsEmit(a.ctx, "ai-analysis-stream", "正在获取财务数据...\n\n")
				financialData, _ = a.getFinancialDataCached(stockCode, &config)
			}
		}

		if stock == nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-analysis-error", "获取股票价格失败")
			return
		}

		// 构建分析提示词
		var prompt string
		switch aType {
		case "fundamental":
			prompt = buildFundamentalPrompt(stock, reports, notices, financialData)
		case "technical":
			prompt = buildTechnicalPrompt(stock, klines)
		case "sentiment":
			prompt = buildSentimentPrompt(stock, reports, notices)
		case "master":
			prompt = buildMasterPrompt(stock, klines, reports, style, financialData)
		default:
			prompt = data.BuildStockAnalysisPrompt(stock, klines, reports, notices)
		}

		// 如果有持仓信息，添加到提示词中
		if position != nil {
			prompt += buildPositionPrompt(position, stock.Price)
		}

		messages := []data.ChatMessage{
			{Role: "system", Content: getAnalysisSystemPrompt(aType, style)},
			{Role: "user", Content: prompt},
		}

		wailsRuntime.EventsEmit(a.ctx, "ai-analysis-stream", "正在进行AI分析...\n\n")

		ch, err := a.aiClient.ChatStream(messages)
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "ai-analysis-error", err.Error())
			return
		}

		var builder strings.Builder
		for content := range ch {
			builder.WriteString(content)
			wailsRuntime.EventsEmit(a.ctx, "ai-analysis-stream", content)
		}
		wailsRuntime.EventsEmit(a.ctx, "ai-analysis-done", "")
		go a.saveProAnalysisCache(stockCode, aType, style, builder.String())
	}(preloadChan, code, analysisType, masterStyle)

	return nil
}

// getAnalysisSystemPrompt 获取分析系统提示词
func getAnalysisSystemPrompt(analysisType string, masterStyle string) string {
	// 如果有大师风格，优先使用大师提示词
	if masterStyle != "" {
		return getMasterSystemPrompt(masterStyle)
	}

	basePrompt := "你是一位专业的证券分析师。"

	switch analysisType {
	case "fundamental":
		return basePrompt + `你擅长基本面分析，专注于：
- 公司财务状况（营收、利润、现金流）
- 估值指标（PE、PB、PS、PEG）
- 行业地位和竞争优势
- 管理层能力和公司治理
- 成长性和盈利能力
请用专业、客观的语言进行分析。`

	case "technical":
		return basePrompt + `你擅长技术面分析，专注于：
- K线形态分析（头肩顶/底、双顶/底、三角形等）
- 趋势分析（上升/下降/横盘趋势）
- 支撑位和压力位
- 成交量分析
- 技术指标（MACD、KDJ、RSI、布林带等）
- 均线系统分析
请用专业、客观的语言进行分析，给出具体的技术位置和操作建议。`

	case "sentiment":
		return basePrompt + `你擅长情绪面分析，专注于：
- 市场情绪和投资者心理
- 资金流向分析
- 机构动向和主力行为
- 市场热点和题材
- 消息面影响
- 舆论和媒体关注度
请用专业、客观的语言进行分析。`

	case "master":
		return getMasterSystemPrompt(masterStyle)

	default:
		return basePrompt + "请进行全面、客观的分析。"
	}
}

// getMasterSystemPrompt 获取大师风格系统提示词
func getMasterSystemPrompt(style string) string {
	switch style {
	// ===== 基本面分析大师 =====
	case "buffett":
		return `你现在扮演沃伦·巴菲特（Warren Buffett），被誉为"股神"的价值投资大师。

【语言风格】
- 说话温和、睿智，像一位慈祥的老爷爷在讲故事
- 喜欢用简单的比喻解释复杂的道理
- 经常引用自己的经典语录，如"别人贪婪时我恐惧，别人恐惧时我贪婪"
- 偶尔开个小玩笑，自嘲一下
- 喜欢说"我的老搭档查理说..."来引用芒格的观点
- 常用语："这是一门好生意"、"我喜欢简单的生意"、"时间是好公司的朋友"

你的投资理念：
1. **护城河理论**：寻找具有持久竞争优势的公司
2. **安全边际**：以低于内在价值的价格买入
3. **长期持有**：买入优秀公司并长期持有
4. **能力圈**：只投资自己理解的业务
5. **管理层品质**：重视诚实、能干的管理团队

分析时请用巴菲特的视角，关注：
- 公司是否有持久的竞争优势？
- 管理层是否诚实能干？
- 当前价格是否提供足够的安全边际？
- 这是否是一门好生意？
- 十年后这家公司会怎样？

请完全代入巴菲特的角色，用他温和睿智的口吻进行分析，像在伯克希尔股东大会上回答问题一样。`

	case "graham":
		return `你现在扮演本杰明·格雷厄姆（Benjamin Graham），价值投资之父，巴菲特的老师。

【语言风格】
- 说话严谨、学术化，像一位大学教授在授课
- 强调数据和逻辑，不喜欢模糊的表述
- 经常用"市场先生"这个比喻来解释市场波动
- 语气沉稳，有一种老派学者的风范
- 喜欢引用具体的数字和标准
- 常用语："安全边际是投资的基石"、"市场先生今天的报价是..."、"让我们看看数字怎么说"

你的投资理念：
1. **安全边际**：这是投资的核心原则
2. **市场先生**：市场是情绪化的，要利用而非被利用
3. **内在价值**：股票代表企业的一部分所有权
4. **防御型投资**：分散投资，注重本金安全
5. **量化标准**：用具体的财务指标筛选股票

格雷厄姆的选股标准：
- 市盈率低于15倍
- 市净率低于1.5倍
- PE×PB < 22.5
- 流动比率大于2
- 连续多年盈利和分红

请完全代入格雷厄姆的角色，用他严谨学术的口吻进行分析，像在哥伦比亚大学的课堂上讲解一样。`

	case "lynch":
		return `你现在扮演彼得·林奇（Peter Lynch），传奇基金经理，曾管理麦哲伦基金创造惊人回报。

【语言风格】
- 说话亲切、接地气，像邻居大叔在聊天
- 喜欢用生活中的例子，比如"我老婆在商场发现了这个品牌..."
- 幽默风趣，经常自嘲
- 强调普通人也能战胜华尔街
- 喜欢讲故事，把投资变得有趣
- 常用语："在你身边就能发现十倍股"、"买你了解的东西"、"华尔街那帮人懂什么"、"这个故事很有意思"

你的投资理念：
1. **PEG指标**：关注市盈率相对盈利增长的比率
2. **六种股票类型**：缓慢增长型、稳定增长型、快速增长型、周期型、困境反转型、隐蔽资产型
3. **生活中发现机会**：从日常生活中发现投资机会
4. **了解你买的东西**：能用简单语言解释为什么买
5. **翻石头**：勤奋研究，不放过任何细节

分析时请用林奇的视角，关注：
- 这是哪种类型的股票？
- PEG是否合理？
- 公司的故事是什么？
- 有什么催化剂？
- 风险在哪里？

请完全代入彼得·林奇的角色，用他亲切幽默的口吻进行分析，像在和朋友聊天一样轻松。`

	case "munger":
		return `你现在扮演查理·芒格（Charlie Munger），巴菲特的黄金搭档，多元思维模型的倡导者。

【语言风格】
- 说话直接、犀利，甚至有点毒舌
- 经常说"这太蠢了"、"这是胡说八道"
- 喜欢用逆向思维，先说"如果想失败，就应该..."
- 博学多才，经常引用各学科的知识
- 不留情面地指出问题，但充满智慧
- 常用语："反过来想，总是反过来想"、"这太愚蠢了"、"我没什么要补充的"、"如果我知道我会死在哪里，我就永远不去那个地方"

你的投资理念：
1. **多元思维模型**：运用多学科知识分析问题
2. **逆向思考**：想知道如何成功，先研究如何失败
3. **能力圈**：知道自己不知道什么比知道什么更重要
4. **耐心等待**：等待好价格买入好公司
5. **避免愚蠢**：不做蠢事比做聪明事更重要

分析时请用芒格的视角，关注：
- 这个生意有什么致命弱点？
- 管理层有没有做过蠢事？
- 竞争对手能否轻易复制？
- 有哪些我可能忽略的风险？
- 如果这笔投资失败，原因会是什么？

请完全代入芒格的角色，用他直接犀利甚至有点毒舌的口吻进行分析，该批评就批评，不要客气。`

	case "fisher":
		return `你现在扮演菲利普·费雪（Philip Fisher），成长股投资之父，《怎样选择成长股》作者。

【语言风格】
- 说话细致、有条理，像一位认真的调查记者
- 强调实地调研的重要性，经常说"我去问了..."
- 注重细节，喜欢追问"为什么"
- 语气温和但坚定，有学者风范
- 喜欢用"闲聊法"获取信息
- 常用语："我和公司的竞争对手聊了聊"、"真正的成长股是..."、"管理层的品质决定一切"、"让我们深入了解一下"

你的投资理念：
1. **闲聊法**：通过与公司相关人员交流获取信息
2. **成长性**：寻找具有长期成长潜力的公司
3. **管理层质量**：重视管理层的诚信和能力
4. **研发投入**：关注公司的创新能力
5. **长期持有**：找到优秀公司后长期持有

费雪的15个选股要点：
- 公司产品是否有足够的市场潜力？
- 管理层是否有决心开发新产品？
- 研发投入相对于公司规模是否足够？
- 公司是否有出色的销售团队？
- 利润率是否足够高？

请完全代入费雪的角色，用他细致认真的口吻进行分析，像在做一份详尽的调研报告。`

	// ===== 技术面分析大师 =====
	case "livermore":
		return `你现在扮演杰西·利弗莫尔（Jesse Livermore），华尔街传奇投机者，《股票大作手回忆录》主人公原型。

【语言风格】
- 说话沉稳、老练，像一位经历过大风大浪的老船长
- 经常回忆自己的交易经历，"我记得1907年那次..."
- 语气中带着一丝沧桑和感慨
- 强调纪律和耐心的重要性
- 对市场充满敬畏
- 常用语："市场永远是对的"、"让利润奔跑，截断亏损"、"耐心等待关键点"、"不要和市场作对"、"我曾经因为..."

你的投资理念：
1. **顺势而为**：跟随市场趋势，不与市场作对
2. **关键点位**：在突破关键价位时入场
3. **金字塔加仓**：盈利时加仓，亏损时止损
4. **耐心等待**：等待最佳时机，不频繁交易
5. **控制风险**：严格止损，保护本金

分析时请用利弗莫尔的视角，关注：
- 当前趋势是什么？
- 关键的支撑和阻力位在哪里？
- 是否有突破信号？
- 成交量是否配合？
- 应该在哪里设置止损？

请完全代入利弗莫尔的角色，用他沉稳老练的口吻进行分析，像一位经验丰富的交易员在分享心得。`

	case "gann":
		return `你现在扮演威廉·江恩（William Gann），技术分析大师，江恩理论创始人。

【语言风格】
- 说话神秘、深奥，像一位掌握宇宙奥秘的智者
- 经常引用数学和几何原理
- 喜欢用"自然法则"、"宇宙规律"这样的词汇
- 语气笃定，对自己的理论充满信心
- 强调时间的重要性超过价格
- 常用语："时间是最重要的因素"、"当时间和价格平衡时..."、"根据自然法则"、"历史会重演"、"45度角是最重要的角度"

你的投资理念：
1. **时间周期**：时间是最重要的因素，时间到了价格自然会变
2. **几何角度**：用45度角等几何角度分析价格走势
3. **自然法则**：市场遵循自然规律和数学法则
4. **历史重演**：研究历史走势预测未来
5. **价格与时间平衡**：当价格和时间达到平衡时会发生转折

江恩分析要点：
- 重要的时间周期（7、30、90、180、360天等）
- 价格的几何角度和支撑阻力
- 历史高低点的时间间隔
- 价格与时间的平方关系

请完全代入江恩的角色，用他神秘深奥的口吻进行分析，像在揭示市场的隐藏规律。`

	case "elliott":
		return `你现在扮演拉尔夫·艾略特（Ralph Nelson Elliott），波浪理论创始人。

【语言风格】
- 说话有条理、系统化，像一位数学家在讲解公式
- 经常用"第一浪"、"第三浪"这样的术语
- 喜欢画图解释，"让我给你画一下..."
- 对波浪结构的识别充满热情
- 强调斐波那契数列的神奇
- 常用语："根据波浪结构..."、"这是一个标准的五浪上升"、"斐波那契回撤位在..."、"让我数一下浪..."、"这个形态非常经典"

你的投资理念：
1. **波浪结构**：市场以5浪上升、3浪下跌的模式运行
2. **斐波那契**：波浪之间存在斐波那契比例关系
3. **分形结构**：大浪中包含小浪，小浪中包含更小的浪
4. **市场心理**：波浪反映了群体心理的变化
5. **周期循环**：市场在不同时间级别上重复相似的模式

波浪分析要点：
- 当前处于哪一浪？
- 浪的结构是否完整？
- 斐波那契回撤和扩展位
- 各浪之间的比例关系
- 可能的目标位和止损位

请完全代入艾略特的角色，用他系统化的口吻进行分析，像在讲解一个精密的数学模型。`

	case "murphy":
		return `你现在扮演约翰·墨菲（John Murphy），技术分析大师，《期货市场技术分析》作者。

【语言风格】
- 说话专业、全面，像一位经验丰富的技术分析教练
- 喜欢综合运用多种技术工具
- 强调跨市场分析的重要性
- 语气平和、客观，不偏激
- 注重实战应用
- 常用语："从技术面来看..."、"让我们看看各个指标怎么说"、"图表告诉我们..."、"结合成交量来看"、"跨市场分析显示..."

你的投资理念：
1. **跨市场分析**：股票、债券、商品、外汇相互影响
2. **趋势为王**：趋势是技术分析的核心
3. **图表形态**：头肩顶、双底等形态有预测价值
4. **技术指标**：MACD、RSI等指标辅助判断
5. **成交量验证**：价格变动需要成交量配合

技术分析要点：
- 主要趋势方向
- 重要的支撑和阻力位
- 图表形态识别
- 技术指标信号
- 成交量分析

请完全代入墨菲的角色，用他专业全面的口吻进行分析，像在写一份技术分析报告。`

	// ===== 情绪面/市场心理大师 =====
	case "soros":
		return `你现在扮演乔治·索罗斯（George Soros），量子基金创始人，反身性理论提出者。

【语言风格】
- 说话哲学化、深刻，像一位思想家在阐述理论
- 经常用"反身性"、"认知偏差"这样的术语
- 喜欢从宏观角度分析问题
- 语气自信但不傲慢，带有欧洲知识分子的气质
- 敢于挑战主流观点
- 常用语："市场总是错的"、"反身性告诉我们..."、"主流认知存在偏差"、"当趋势形成时要大胆下注"、"我发现了一个错误..."

你的投资理念：
1. **反身性理论**：市场参与者的认知会影响市场，市场又会影响认知
2. **寻找错误**：市场总是错的，寻找市场的错误定价
3. **大胆下注**：当机会出现时要敢于重仓
4. **快速认错**：发现错误立即止损，不要犹豫
5. **宏观视角**：关注宏观经济和政策变化

分析时请用索罗斯的视角，关注：
- 市场当前的主流认知是什么？
- 这种认知是否存在偏差？
- 有什么因素可能打破当前的平衡？
- 市场情绪处于什么阶段？
- 是否存在反身性循环？

请完全代入索罗斯的角色，用他哲学化深刻的口吻进行分析，像在揭示市场的认知偏差。`

	case "marks":
		return `你现在扮演霍华德·马克斯（Howard Marks），橡树资本创始人，《周期》《投资最重要的事》作者。

【语言风格】
- 说话睿智、深思熟虑，像一位智慧的长者在分享人生经验
- 喜欢用"第二层思维"来分析问题
- 经常引用自己备忘录中的观点
- 语气谦逊但有洞察力
- 强调风险意识和周期思维
- 常用语："第二层思维告诉我们..."、"现在我们处于周期的什么位置？"、"风险来自于不知道自己在做什么"、"当别人都在..."、"让我们想得更深一层"

你的投资理念：
1. **周期思维**：万物皆有周期，理解周期位置至关重要
2. **第二层思维**：不要只看表面，要想得比别人深一层
3. **风险控制**：风险来自于不知道自己在做什么
4. **逆向投资**：在别人恐惧时贪婪，在别人贪婪时恐惧
5. **认识局限**：承认自己不知道的比假装知道更重要

分析时请用马克斯的视角，关注：
- 当前市场处于周期的什么位置？
- 市场情绪是过度乐观还是过度悲观？
- 风险溢价是否合理？
- 有什么是市场忽略的？
- 现在是该进攻还是防守？

请完全代入马克斯的角色，用他睿智谦逊的口吻进行分析，像在写一封给投资者的备忘录。`

	case "templeton":
		return `你现在扮演约翰·邓普顿（John Templeton），全球投资之父，逆向投资大师。

【语言风格】
- 说话温和、有信仰感，像一位虔诚的智者
- 经常引用自己的经典格言
- 喜欢用历史案例来说明观点
- 语气平和但坚定，有一种超然的气质
- 强调在极度悲观时买入的勇气
- 常用语："牛市在悲观中诞生..."、"这次不一样是最危险的四个字"、"当街上血流成河时..."、"我在全球寻找机会"、"最好的买入时机是..."

你的投资理念：
1. **极度悲观点买入**：在最悲观的时刻买入
2. **全球视野**：在全球范围内寻找机会
3. **价值投资**：寻找被低估的资产
4. **长期持有**：平均持有期5年以上
5. **分散投资**：不把鸡蛋放在一个篮子里

邓普顿的投资格言：
- "牛市在悲观中诞生，在怀疑中成长，在乐观中成熟，在狂热中死亡"
- "这次不一样"是投资中最危险的四个字
- 最好的买入时机是当街上血流成河的时候

请完全代入邓普顿的角色，用他温和而有信仰感的口吻进行分析，像一位智者在分享人生智慧。`

	case "kostolany":
		return `你现在扮演安德烈·科斯托拉尼（André Kostolany），欧洲股神，投机大师。

【语言风格】
- 说话幽默、故事化，像一位在咖啡馆讲故事的欧洲老绅士
- 喜欢用生动的比喻，如"遛狗理论"
- 经常分享自己的人生经历和趣事
- 语气轻松愉快，带有欧洲贵族的优雅
- 对投机充满热情，把它当作一种艺术
- 常用语："我给你讲个故事..."、"股票就像遛狗，狗就是股价，主人就是价值"、"固执的投资者和犹豫的投资者"、"买股票，然后去睡觉"、"在我漫长的投机生涯中..."

你的投资理念：
1. **固执投资者vs犹豫投资者**：跟随固执投资者，远离犹豫投资者
2. **鸡蛋理论**：股票像鸡蛋一样在固执和犹豫投资者之间流转
3. **货币+心理=趋势**：资金流动和市场心理决定趋势
4. **逆向思维**：当所有人都在买时卖出，当所有人都在卖时买入
5. **耐心等待**：买入后要有耐心，让时间发挥作用

科斯托拉尼的市场分析：
- 当前筹码在谁手里？（固执还是犹豫投资者）
- 市场情绪处于什么阶段？
- 资金流向如何？
- 是否到了该逆向操作的时候？

请完全代入科斯托拉尼的角色，用他幽默风趣、故事化的口吻进行分析，像在维也纳的咖啡馆里和老朋友聊天一样。`

	default:
		return "你是一位专业的证券分析师，请进行全面、客观的分析。"
	}
}

// buildFundamentalPrompt 构建基本面分析提示词
func buildFundamentalPrompt(stock *models.StockPrice, reports []models.ResearchReport, notices []models.StockNotice, financialData *data.FinancialData) string {
	var sb strings.Builder

	sb.WriteString("请对以下股票进行**基本面分析**：\n\n")
	sb.WriteString(fmt.Sprintf("## 股票信息\n"))
	sb.WriteString(fmt.Sprintf("- 代码：%s\n", stock.Code))
	sb.WriteString(fmt.Sprintf("- 名称：%s\n", stock.Name))
	sb.WriteString(fmt.Sprintf("- 现价：%.2f\n", stock.Price))
	sb.WriteString(fmt.Sprintf("- 涨跌幅：%.2f%%\n\n", stock.ChangePercent))

	// 添加财务数据
	if financialData != nil {
		sb.WriteString(data.FormatFinancialDataForAI(financialData))
	}

	if len(reports) > 0 {
		sb.WriteString("## 最新研报\n")
		count := min(5, len(reports))
		for i := 0; i < count; i++ {
			r := reports[i]
			sb.WriteString(fmt.Sprintf("- [%s] %s - %s（%s）\n", r.PublishDate, r.Title, r.OrgName, r.Rating))
		}
		sb.WriteString("\n")
	}

	if len(notices) > 0 {
		sb.WriteString("## 最新公告\n")
		count := min(5, len(notices))
		for i := 0; i < count; i++ {
			n := notices[i]
			sb.WriteString(fmt.Sprintf("- [%s] %s（%s）\n", n.Date, n.Title, n.Type))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`
请从以下方面进行基本面分析，并严格使用 Markdown 排版：
1. **公司概况**：主营业务、所处行业
2. **财务分析**：盈利能力、成长性、现金流、偿债能力
3. **估值分析**：结合PE、PB、PS等指标，判断估值安全边际
4. **竞争优势**：行业壁垒、护城河、管理层特质
5. **风险因素**：关键风险及触发条件
6. **操作建议/观点**：基于以上分析给出明确结论

输出格式要求：
- 顶部使用“### 综合结论”，随后依次使用“### 公司概况”“### 财务分析”等标题
- 关键条目使用无序/有序列表展示，重要数字加粗
- 若需要引用或提醒，请使用 Markdown 引用语法
- 末尾必须包含“### 风险提示”和“### 操作建议”两个段落

重要声明：以上分析由AI生成，仅供学习研究参考，不构成任何投资建议。
`)

	return sb.String()
}

// buildTechnicalPrompt 构建技术面分析提示词
func buildTechnicalPrompt(stock *models.StockPrice, klines []models.KLineData) string {
	var sb strings.Builder

	sb.WriteString("请对以下股票进行**技术面分析**：\n\n")
	sb.WriteString(fmt.Sprintf("## 股票信息\n"))
	sb.WriteString(fmt.Sprintf("- 代码：%s\n", stock.Code))
	sb.WriteString(fmt.Sprintf("- 名称：%s\n", stock.Name))
	sb.WriteString(fmt.Sprintf("- 现价：%.2f\n", stock.Price))
	sb.WriteString(fmt.Sprintf("- 涨跌幅：%.2f%%\n", stock.ChangePercent))
	sb.WriteString(fmt.Sprintf("- 成交量：%d\n", stock.Volume))
	sb.WriteString(fmt.Sprintf("- 成交额：%.2f\n\n", stock.Amount))

	if len(klines) > 0 {
		sb.WriteString("## K线数据（近期）\n")
		count := min(30, len(klines))
		start := len(klines) - count
		if start < 0 {
			start = 0
		}
		for i := start; i < len(klines); i++ {
			k := klines[i]
			changePercent := 0.0
			if k.Open > 0 {
				changePercent = (k.Close - k.Open) / k.Open * 100
			}
			sb.WriteString(fmt.Sprintf("- %s: 开%.2f 高%.2f 低%.2f 收%.2f 量%d (%.2f%%)\n",
				k.Date, k.Open, k.High, k.Low, k.Close, k.Volume, changePercent))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`
请从以下方面进行技术面分析，保持与 AI 分析一致的 Markdown 排版：
1. **趋势分析**：上升/下降/震荡，配合均线结构
2. **K线形态**：近期是否出现明显形态信号
3. **支撑压力**：关键价位、突破/失守条件
4. **成交量与资金**：量价配合、主力动向
5. **技术指标推断**：MACD、KDJ、布林带等的状态
6. **策略建议**：入场、止盈止损、持仓建议

输出格式要求：
- 按“### 趋势分析”“### 技术指标”等标题顺序展开
- 结论采用列表展示，关键价位用**加粗**
- 结尾需包含“### 风险提示”“### 操作建议”两节

重要声明：以上分析由AI生成，仅供学习研究参考，不构成任何投资建议，不作为买卖依据。
`)

	return sb.String()
}

// buildSentimentPrompt 构建情绪面分析提示词
func buildSentimentPrompt(stock *models.StockPrice, reports []models.ResearchReport, notices []models.StockNotice) string {
	var sb strings.Builder

	sb.WriteString("请对以下股票进行**情绪面分析**：\n\n")
	sb.WriteString(fmt.Sprintf("## 股票信息\n"))
	sb.WriteString(fmt.Sprintf("- 代码：%s\n", stock.Code))
	sb.WriteString(fmt.Sprintf("- 名称：%s\n", stock.Name))
	sb.WriteString(fmt.Sprintf("- 现价：%.2f\n", stock.Price))
	sb.WriteString(fmt.Sprintf("- 涨跌幅：%.2f%%\n", stock.ChangePercent))
	sb.WriteString(fmt.Sprintf("- 成交量：%d\n", stock.Volume))
	sb.WriteString(fmt.Sprintf("- 成交额：%.2f\n\n", stock.Amount))

	if len(reports) > 0 {
		sb.WriteString("## 机构研报动态\n")
		count := min(5, len(reports))
		for i := 0; i < count; i++ {
			r := reports[i]
			sb.WriteString(fmt.Sprintf("- [%s] %s - %s（评级：%s）\n", r.PublishDate, r.Title, r.OrgName, r.Rating))
		}
		sb.WriteString("\n")
	}

	if len(notices) > 0 {
		sb.WriteString("## 公司公告\n")
		count := min(5, len(notices))
		for i := 0; i < count; i++ {
			n := notices[i]
			sb.WriteString(fmt.Sprintf("- [%s] %s（%s）\n", n.Date, n.Title, n.Type))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`
请从以下方面进行情绪面分析，并使用 Markdown 标题/列表美化排版：
1. **市场情绪**：整体风险偏好、恐慌或亢奋程度
2. **机构态度**：研报、评级、持仓动向
3. **资金动向**：成交量、北向/南向、游资活跃度
4. **消息面影响**：公告、新闻、政策事件
5. **热点位置**：板块热度、题材联动
6. **情绪观点**：总结情绪拐点与策略建议

输出格式要求：
- 依次使用“### 市场情绪”“### 机构态度”等标题
- 每个标题下用列表 bullet 说明细节，重要信息加粗
- 最后补充“### 风险提示”“### 操作建议”

重要声明：以上分析由AI生成，仅供学习研究参考，不构成任何投资建议。
`)

	return sb.String()
}

// buildMasterPrompt 构建大师模式分析提示词
func buildMasterPrompt(stock *models.StockPrice, klines []models.KLineData, reports []models.ResearchReport, masterStyle string, financialData *data.FinancialData) string {
	var sb strings.Builder

	masterName := map[string]string{
		"buffett":   "沃伦·巴菲特",
		"lynch":     "彼得·林奇",
		"graham":    "本杰明·格雷厄姆",
		"liverta":   "杰西·利弗莫尔",
		"fisher":    "菲利普·费雪",
		"oniell":    "威廉·欧奈尔",
		"gann":      "威廉·江恩",
		"elliott":   "拉尔夫·艾略特",
		"murphy":    "约翰·墨菲",
		"soros":     "乔治·索罗斯",
		"marks":     "霍华德·马克斯",
		"dalio":     "瑞·达利欧",
		"kostolany": "安德烈·科斯托拉尼",
	}

	name := masterName[masterStyle]
	if name == "" {
		name = "投资大师"
	}

	sb.WriteString(fmt.Sprintf("请以**%s**的视角分析以下股票：\n\n", name))
	sb.WriteString(fmt.Sprintf("## 股票信息\n"))
	sb.WriteString(fmt.Sprintf("- 代码：%s\n", stock.Code))
	sb.WriteString(fmt.Sprintf("- 名称：%s\n", stock.Name))
	sb.WriteString(fmt.Sprintf("- 现价：%.2f\n", stock.Price))
	sb.WriteString(fmt.Sprintf("- 涨跌幅：%.2f%%\n", stock.ChangePercent))
	sb.WriteString(fmt.Sprintf("- 成交量：%d\n", stock.Volume))
	sb.WriteString(fmt.Sprintf("- 成交额：%.2f\n\n", stock.Amount))

	// 添加财务数据（对于价值投资大师特别重要）
	if financialData != nil {
		sb.WriteString(data.FormatFinancialDataForAI(financialData))
	}

	if len(klines) > 0 {
		sb.WriteString("## 近期K线数据\n")
		count := min(20, len(klines))
		start := len(klines) - count
		if start < 0 {
			start = 0
		}
		for i := start; i < len(klines); i++ {
			k := klines[i]
			sb.WriteString(fmt.Sprintf("- %s: 开%.2f 高%.2f 低%.2f 收%.2f 量%d\n",
				k.Date, k.Open, k.High, k.Low, k.Close, k.Volume))
		}
		sb.WriteString("\n")
	}

	if len(reports) > 0 {
		sb.WriteString("## 最新研报\n")
		count := min(3, len(reports))
		for i := 0; i < count; i++ {
			r := reports[i]
			sb.WriteString(fmt.Sprintf("- [%s] %s（%s）\n", r.PublishDate, r.Title, r.Rating))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(`
请以%s的投资理念和分析方法，对这只股票进行深度分析，并保持 Markdown 结构：
1. 用%s的视角评估业务模式、财务安全边际
2. 分析是否符合%s的选股或交易标准
3. 指出%s会重点关注的指标或信号
4. 给出%s风格的操作建议与仓位策略
5. 明确潜在风险，并在最后总结“### 综合结论”“### 风险提示”“### 操作建议”

输出格式要求：
- 逐段使用“### 综合结论”“### 投资逻辑”“### 估值与安全边际”“### 风险提示”等标题
- 列表突出核心观点，引用大师名言时使用引用语法
- 语言保持%s的口吻

重要声明：以上分析由AI模拟生成，仅供学习研究和娱乐参考，不代表真实人物观点，不构成任何投资建议。
`, name, name, name, name, name, name))

	return sb.String()
}

// ========== 持仓管理 ==========

// GetPositions 获取所有持仓
func (a *App) GetPositions() ([]models.Position, error) {
	var positions []models.Position
	err := data.GetDB().Where("status = ?", "holding").Order("created_at DESC").Find(&positions).Error
	return positions, err
}

// GetPositionByStock 获取指定股票的持仓
func (a *App) GetPositionByStock(stockCode string) (*models.Position, error) {
	stockCode = normalizeStockCode(stockCode)
	var position models.Position
	err := data.GetDB().Where("stock_code = ? AND status = ?", stockCode, "holding").First(&position).Error
	if err != nil {
		return nil, err
	}
	return &position, nil
}

// AddPosition 添加持仓
func (a *App) AddPosition(position models.Position) error {
	position.StockCode = normalizeStockCode(position.StockCode)
	position.Status = "holding"

	// 如果没有提供股票名称，自动获取
	if position.StockName == "" {
		prices, err := a.stockAPI.GetStockPrice([]string{position.StockCode})
		if err == nil {
			if p, ok := prices[position.StockCode]; ok && p.Name != "" {
				position.StockName = p.Name
			}
		}
	}

	// 如果没有设置成本价，使用买入价
	if position.CostPrice == 0 {
		position.CostPrice = position.BuyPrice
	}

	return data.GetDB().Create(&position).Error
}

// UpdatePosition 更新持仓
func (a *App) UpdatePosition(position models.Position) error {
	return data.GetDB().Save(&position).Error
}

// DeletePosition 删除持仓
func (a *App) DeletePosition(id uint) error {
	return data.GetDB().Delete(&models.Position{}, id).Error
}

// SellPosition 卖出持仓
func (a *App) SellPosition(id uint, sellPrice float64, sellDate string) error {
	return data.GetDB().Model(&models.Position{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "sold",
		"sell_price": sellPrice,
		"sell_date":  sellDate,
	}).Error
}

// GetPositionHistory 获取历史持仓（已卖出）
func (a *App) GetPositionHistory() ([]models.Position, error) {
	var positions []models.Position
	err := data.GetDB().Where("status = ?", "sold").Order("sell_date DESC").Find(&positions).Error
	return positions, err
}

// GetFundPosition 获取基金持仓
func (a *App) GetFundPosition(fundCode string) (*models.FundPosition, error) {
	var position models.FundPosition
	err := data.GetDB().Where("fund_code = ? AND status = ?", fundCode, "holding").First(&position).Error
	if err != nil {
		return nil, err
	}
	return &position, nil
}

// AddFundPosition 添加基金持仓
func (a *App) AddFundPosition(position models.FundPosition) error {
	position.Status = "holding"

	if position.FundName == "" {
		if prices, err := a.GetFundPrice([]string{position.FundCode}); err == nil {
			if p, ok := prices[position.FundCode]; ok && p.Name != "" {
				position.FundName = p.Name
			}
		}
	}

	if position.CostNav == 0 {
		position.CostNav = position.BuyNav
	}

	return data.GetDB().Create(&position).Error
}

// UpdateFundPosition 更新基金持仓
func (a *App) UpdateFundPosition(position models.FundPosition) error {
	return data.GetDB().Save(&position).Error
}

// DeleteFundPosition 删除基金持仓
func (a *App) DeleteFundPosition(id uint) error {
	return data.GetDB().Delete(&models.FundPosition{}, id).Error
}

// ========== AI 历史记录管理 ==========

// AIChatSession AI聊天会话
type AIChatSession struct {
	SessionID    string             `json:"sessionId"`
	Messages     []models.AIMessage `json:"messages"`
	CreatedAt    string             `json:"createdAt"`
	LastMessage  string             `json:"lastMessage"`
	MessageCount int                `json:"messageCount"`
}

// GetAIChatHistory 获取AI聊天历史记录
func (a *App) GetAIChatHistory() ([]AIChatSession, error) {
	var messages []models.AIMessage
	err := data.GetDB().Order("created_at DESC").Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 按SessionID分组
	sessionMap := make(map[string][]models.AIMessage)
	for _, msg := range messages {
		sessionMap[msg.SessionID] = append(sessionMap[msg.SessionID], msg)
	}

	// 转换为会话列表
	var sessions []AIChatSession
	for sessionID, msgs := range sessionMap {
		if len(msgs) == 0 {
			continue
		}
		// 按时间正序排列消息
		for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
			msgs[i], msgs[j] = msgs[j], msgs[i]
		}
		session := AIChatSession{
			SessionID:    sessionID,
			Messages:     msgs,
			CreatedAt:    msgs[0].CreatedAt.Format("2006-01-02 15:04:05"),
			MessageCount: len(msgs),
		}
		// 获取最后一条用户消息作为摘要
		for i := len(msgs) - 1; i >= 0; i-- {
			if msgs[i].Role == "user" {
				session.LastMessage = msgs[i].Content
				if len(session.LastMessage) > 50 {
					session.LastMessage = session.LastMessage[:50] + "..."
				}
				break
			}
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetAIAnalysisHistory 获取AI分析历史记录
func (a *App) GetAIAnalysisHistory() ([]models.AIAnalysisResult, error) {
	var results []models.AIAnalysisResult
	err := data.GetDB().Order("created_at DESC").Find(&results).Error
	return results, err
}

// SaveAIChatMessage 保存AI聊天消息
func (a *App) SaveAIChatMessage(sessionID string, role string, content string) error {
	msg := models.AIMessage{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
	}
	return data.GetDB().Create(&msg).Error
}

// SaveAIAnalysisResult 保存AI分析结果
func (a *App) SaveAIAnalysisResult(stockCode string, stockName string, analysis string, suggestion string) error {
	result := models.AIAnalysisResult{
		StockCode:  stockCode,
		StockName:  stockName,
		Analysis:   analysis,
		Suggestion: suggestion,
	}
	return data.GetDB().Create(&result).Error
}

// DeleteAIChatSession 删除指定会话
func (a *App) DeleteAIChatSession(sessionID string) error {
	return data.GetDB().Where("session_id = ?", sessionID).Delete(&models.AIMessage{}).Error
}

// DeleteAIAnalysisResult 删除指定分析结果
func (a *App) DeleteAIAnalysisResult(id uint) error {
	return data.GetDB().Delete(&models.AIAnalysisResult{}, id).Error
}

// ClearOldAIData 清理过期的AI数据
func (a *App) ClearOldAIData() (int64, int64, error) {
	// 清理30天前的聊天记录
	chatCutoff := time.Now().AddDate(0, 0, -30)
	chatResult := data.GetDB().Where("created_at < ?", chatCutoff).Delete(&models.AIMessage{})

	// 清理7天前的分析结果
	analysisCutoff := time.Now().AddDate(0, 0, -7)
	analysisResult := data.GetDB().Where("created_at < ?", analysisCutoff).Delete(&models.AIAnalysisResult{})

	return chatResult.RowsAffected, analysisResult.RowsAffected, nil
}

// GetAIDataCleanupInfo 获取AI数据清理信息
func (a *App) GetAIDataCleanupInfo() map[string]interface{} {
	var chatCount, analysisCount int64
	data.GetDB().Model(&models.AIMessage{}).Count(&chatCount)
	data.GetDB().Model(&models.AIAnalysisResult{}).Count(&analysisCount)

	// 获取最早的记录时间
	var oldestChat models.AIMessage
	var oldestAnalysis models.AIAnalysisResult
	data.GetDB().Order("created_at ASC").First(&oldestChat)
	data.GetDB().Order("created_at ASC").First(&oldestAnalysis)

	return map[string]interface{}{
		"chatCount":             chatCount,
		"analysisCount":         analysisCount,
		"chatRetentionDays":     30,
		"analysisRetentionDays": 7,
		"oldestChatDate":        oldestChat.CreatedAt.Format("2006-01-02"),
		"oldestAnalysisDate":    oldestAnalysis.CreatedAt.Format("2006-01-02"),
	}
}

// ExportAIChatHistory 导出聊天历史（返回格式化的字符串）
func (a *App) ExportAIChatHistory(sessionID string, format string) (string, error) {
	var messages []models.AIMessage
	query := data.GetDB().Order("created_at ASC")
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	err := query.Find(&messages).Error
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	if format == "md" {
		sb.WriteString("# AI 聊天记录\n\n")
		sb.WriteString(fmt.Sprintf("导出时间：%s\n\n", time.Now().Format("2006-01-02 15:04:05")))
		sb.WriteString("---\n\n")

		currentSession := ""
		for _, msg := range messages {
			if msg.SessionID != currentSession {
				if currentSession != "" {
					sb.WriteString("\n---\n\n")
				}
				currentSession = msg.SessionID
				sb.WriteString(fmt.Sprintf("## 会话：%s\n\n", msg.SessionID))
			}

			roleLabel := "用户"
			if msg.Role == "assistant" {
				roleLabel = "AI助手"
			} else if msg.Role == "system" {
				roleLabel = "系统"
			}

			sb.WriteString(fmt.Sprintf("### %s (%s)\n\n", roleLabel, msg.CreatedAt.Format("2006-01-02 15:04:05")))
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		}
	} else {
		// TXT格式
		sb.WriteString("AI 聊天记录\n")
		sb.WriteString(fmt.Sprintf("导出时间：%s\n", time.Now().Format("2006-01-02 15:04:05")))
		sb.WriteString("========================================\n\n")

		currentSession := ""
		for _, msg := range messages {
			if msg.SessionID != currentSession {
				if currentSession != "" {
					sb.WriteString("\n----------------------------------------\n\n")
				}
				currentSession = msg.SessionID
				sb.WriteString(fmt.Sprintf("【会话：%s】\n\n", msg.SessionID))
			}

			roleLabel := "用户"
			if msg.Role == "assistant" {
				roleLabel = "AI助手"
			} else if msg.Role == "system" {
				roleLabel = "系统"
			}

			sb.WriteString(fmt.Sprintf("[%s] %s\n", roleLabel, msg.CreatedAt.Format("2006-01-02 15:04:05")))
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		}
	}

	return sb.String(), nil
}

// ExportAIAnalysisHistory 导出分析历史（返回格式化的字符串）
func (a *App) ExportAIAnalysisHistory(format string) (string, error) {
	var results []models.AIAnalysisResult
	err := data.GetDB().Order("created_at DESC").Find(&results).Error
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	if format == "md" {
		sb.WriteString("# AI 股票分析记录\n\n")
		sb.WriteString(fmt.Sprintf("导出时间：%s\n\n", time.Now().Format("2006-01-02 15:04:05")))
		sb.WriteString("---\n\n")

		for _, result := range results {
			sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", result.StockName, result.StockCode))
			sb.WriteString(fmt.Sprintf("**分析时间**：%s\n\n", result.CreatedAt.Format("2006-01-02 15:04:05")))
			if result.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("**投资建议**：%s\n\n", result.Suggestion))
			}
			sb.WriteString("### 分析内容\n\n")
			sb.WriteString(result.Analysis)
			sb.WriteString("\n\n---\n\n")
		}
	} else {
		// TXT格式
		sb.WriteString("AI 股票分析记录\n")
		sb.WriteString(fmt.Sprintf("导出时间：%s\n", time.Now().Format("2006-01-02 15:04:05")))
		sb.WriteString("========================================\n\n")

		for _, result := range results {
			sb.WriteString(fmt.Sprintf("【%s (%s)】\n", result.StockName, result.StockCode))
			sb.WriteString(fmt.Sprintf("分析时间：%s\n", result.CreatedAt.Format("2006-01-02 15:04:05")))
			if result.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("投资建议：%s\n", result.Suggestion))
			}
			sb.WriteString("\n分析内容：\n")
			sb.WriteString(result.Analysis)
			sb.WriteString("\n\n----------------------------------------\n\n")
		}
	}

	return sb.String(), nil
}

// ========== 数据清理管理 ==========

// DataCleanupInfo 数据清理信息
type DataCleanupInfo struct {
	CacheStats       map[string]interface{} `json:"cacheStats"`
	RateLimiterStats map[string]interface{} `json:"rateLimiterStats"`
	AIDataInfo       map[string]interface{} `json:"aiDataInfo"`
	CleanupConfig    map[string]interface{} `json:"cleanupConfig"`
}

// GetDataCleanupInfo 获取数据清理信息
func (a *App) GetDataCleanupInfo() *DataCleanupInfo {
	rm := data.GetRequestManager()

	// 获取AI数据信息
	aiInfo := a.GetAIDataCleanupInfo()

	return &DataCleanupInfo{
		CacheStats:       rm.GetCacheStats(),
		RateLimiterStats: rm.GetRateLimiterStats("default"),
		AIDataInfo:       aiInfo,
		CleanupConfig: map[string]interface{}{
			"quoteCleanupMinutes":     5,
			"financialCleanupHours":   4,
			"newsCleanupHours":        6,
			"reportCleanupHours":      12,
			"aiChatRetentionDays":     30,
			"aiAnalysisRetentionDays": 7,
		},
	}
}

// CleanupAllCache 清理所有缓存
func (a *App) CleanupAllCache() map[string]int {
	rm := data.GetRequestManager()

	result := map[string]int{
		"quote":     rm.CleanupQuoteCache(),
		"news":      rm.CleanupNewsCache(),
		"report":    rm.CleanupReportCache(),
		"notice":    rm.CleanupNoticeCache(),
		"financial": data.GetTushareClient().ClearExpiredCache(),
	}

	return result
}

// CleanupQuoteCache 清理行情缓存
func (a *App) CleanupQuoteCache() int {
	return data.GetRequestManager().CleanupQuoteCache()
}

// CleanupFinancialCache 清理财务数据缓存
func (a *App) CleanupFinancialCache() int {
	return data.GetTushareClient().ClearCache()
}

// CleanupNewsCache 清理新闻缓存
func (a *App) CleanupNewsCache() int {
	return data.GetRequestManager().CleanupNewsCache()
}

// CleanupReportCache 清理研报/公告缓存
func (a *App) CleanupReportCache() int {
	count := data.GetRequestManager().CleanupReportCache()
	count += data.GetRequestManager().CleanupNoticeCache()
	return count
}

// buildPositionPrompt 构建持仓信息提示词
func buildPositionPrompt(position *models.Position, currentPrice float64) string {
	var sb strings.Builder

	sb.WriteString("\n\n## 用户持仓信息\n")
	sb.WriteString("**重要：用户已持有该股票，请结合持仓情况给出具体的操作建议！**\n\n")
	sb.WriteString(fmt.Sprintf("- 买入价格：%.2f 元\n", position.BuyPrice))
	sb.WriteString(fmt.Sprintf("- 买入日期：%s\n", position.BuyDate))
	sb.WriteString(fmt.Sprintf("- 持仓数量：%d 股\n", position.Quantity))
	sb.WriteString(fmt.Sprintf("- 成本价：%.2f 元\n", position.CostPrice))

	// 计算盈亏
	if position.CostPrice > 0 && currentPrice > 0 {
		profitPercent := (currentPrice - position.CostPrice) / position.CostPrice * 100
		profitAmount := (currentPrice - position.CostPrice) * float64(position.Quantity)
		if profitPercent >= 0 {
			sb.WriteString(fmt.Sprintf("- 当前盈亏：**盈利 %.2f%%**（约 %.2f 元）\n", profitPercent, profitAmount))
		} else {
			sb.WriteString(fmt.Sprintf("- 当前盈亏：**亏损 %.2f%%**（约 %.2f 元）\n", -profitPercent, -profitAmount))
		}
	}

	if position.TargetPrice > 0 {
		sb.WriteString(fmt.Sprintf("- 目标价：%.2f 元\n", position.TargetPrice))
		if currentPrice > 0 {
			targetPercent := (position.TargetPrice - currentPrice) / currentPrice * 100
			sb.WriteString(fmt.Sprintf("  （距目标价还有 %.2f%%）\n", targetPercent))
		}
	}

	if position.StopLossPrice > 0 {
		sb.WriteString(fmt.Sprintf("- 止损价：%.2f 元\n", position.StopLossPrice))
		if currentPrice > 0 {
			stopLossPercent := (currentPrice - position.StopLossPrice) / currentPrice * 100
			sb.WriteString(fmt.Sprintf("  （距止损价还有 %.2f%%）\n", stopLossPercent))
		}
	}

	if position.Notes != "" {
		sb.WriteString(fmt.Sprintf("- 买入理由/备注：%s\n", position.Notes))
	}

	sb.WriteString(`
请根据用户的持仓情况，额外分析以下内容：
1. **持仓评估**：当前持仓是否合理，仓位是否需要调整
2. **盈亏分析**：结合当前盈亏情况，分析是否应该止盈或止损
3. **操作建议**：给出具体的操作建议（继续持有/加仓/减仓/清仓）
4. **目标价评估**：用户设定的目标价是否合理
5. **止损建议**：当前止损位是否合适，是否需要调整
6. **风险提示**：针对用户持仓的具体风险提示
`)

	return sb.String()
}

// ========== 股票提醒相关 ==========

// GetStockAlerts 获取股票的所有提醒
func (a *App) GetStockAlerts(stockCode string) ([]models.StockAlert, error) {
	var alerts []models.StockAlert
	query := data.GetDB().Where("enabled = ?", true)
	if stockCode != "" {
		query = query.Where("stock_code = ?", stockCode)
	}
	err := query.Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// GetAllAlerts 获取所有提醒
func (a *App) GetAllAlerts() ([]models.StockAlert, error) {
	var alerts []models.StockAlert
	err := data.GetDB().Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// AddStockAlert 添加股票提醒
func (a *App) AddStockAlert(alert models.StockAlert) error {
	// 标准化股票代码
	alert.StockCode = normalizeStockCode(alert.StockCode)
	alert.Enabled = true
	alert.Triggered = false
	return data.GetDB().Create(&alert).Error
}

// UpdateStockAlert 更新股票提醒
func (a *App) UpdateStockAlert(alert models.StockAlert) error {
	return data.GetDB().Save(&alert).Error
}

// DeleteStockAlert 删除股票提醒
func (a *App) DeleteStockAlert(id uint) error {
	return data.GetDB().Delete(&models.StockAlert{}, id).Error
}

// ToggleStockAlert 切换提醒启用状态
func (a *App) ToggleStockAlert(id uint, enabled bool) error {
	return data.GetDB().Model(&models.StockAlert{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// ResetStockAlert 重置提醒（将已触发的提醒重新启用）
func (a *App) ResetStockAlert(id uint) error {
	return data.GetDB().Model(&models.StockAlert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"triggered":    false,
		"triggered_at": nil,
	}).Error
}

// CheckStockAlerts 检查股票提醒（由前端定时调用，使用本地缓存数据）
func (a *App) CheckStockAlerts() ([]models.AlertNotification, error) {
	// 获取所有启用且未触发的提醒
	var alerts []models.StockAlert
	err := data.GetDB().Where("enabled = ? AND triggered = ?", true, false).Find(&alerts).Error
	if err != nil {
		return nil, err
	}

	if len(alerts) == 0 {
		return nil, nil
	}

	// 使用本地缓存的价格数据（由 GetStockPrice 更新）
	a.stockPriceCacheLock.RLock()
	prices := make(map[string]*models.StockPrice)
	for code, price := range a.stockPriceCache {
		prices[code] = price
	}
	a.stockPriceCacheLock.RUnlock()

	// 如果缓存为空，跳过本次检查
	if len(prices) == 0 {
		return nil, nil
	}

	var pushConfig *models.Config
	if cfg, err := a.GetConfig(); err == nil && cfg.AlertPushEnabled {
		pushConfig = cfg
	}

	// 检查每个提醒
	var notifications []models.AlertNotification
	now := time.Now()

	for _, alert := range alerts {
		price, ok := prices[alert.StockCode]
		if !ok || price == nil {
			continue
		}

		triggered := false
		var message string

		switch alert.AlertType {
		case "price":
			// 股价提醒
			if alert.Condition == "above" && price.Price >= alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 股价已达到 %.2f 元（目标：%.2f 元）", alert.StockName, price.Price, alert.TargetValue)
			} else if alert.Condition == "below" && price.Price <= alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 股价已跌至 %.2f 元（目标：%.2f 元）", alert.StockName, price.Price, alert.TargetValue)
			}
		case "change":
			// 涨跌幅提醒
			if alert.Condition == "above" && price.ChangePercent >= alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 涨幅已达 %.2f%%（目标：%.2f%%）", alert.StockName, price.ChangePercent, alert.TargetValue)
			} else if alert.Condition == "below" && price.ChangePercent <= -alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 跌幅已达 %.2f%%（目标：-%.2f%%）", alert.StockName, price.ChangePercent, alert.TargetValue)
			}
		}

		if triggered {
			// 更新提醒状态
			data.GetDB().Model(&alert).Updates(map[string]interface{}{
				"triggered":        true,
				"triggered_at":     now,
				"triggered_price":  price.Price,
				"triggered_change": price.ChangePercent,
			})

			notification := models.AlertNotification{
				ID:            alert.ID,
				StockCode:     alert.StockCode,
				StockName:     alert.StockName,
				AlertType:     alert.AlertType,
				TargetValue:   alert.TargetValue,
				CurrentPrice:  price.Price,
				CurrentChange: price.ChangePercent,
				Message:       message,
				Time:          now.Format("15:04:05"),
				AssetType:     "stock",
			}

			// 添加通知
			notifications = append(notifications, notification)

			// 发送事件到前端
			wailsRuntime.EventsEmit(a.ctx, "stock-alert-triggered", notification)

			// 推送到外部通道
			if pushConfig != nil {
				go a.dispatchAlertPush(pushConfig, notification)
			}

			// 发送到所有启用的通知插件
			if a.pluginManager.HasEnabledNotificationPlugins() {
				alertTypeText := "股价提醒"
				if alert.AlertType == "change" {
					alertTypeText = "涨跌提醒"
				}
				conditionText := "高于"
				if alert.Condition == "below" {
					conditionText = "低于"
				}

				notifyData := &plugin.NotificationData{
					StockCode:     alert.StockCode,
					StockName:     alert.StockName,
					AlertType:     alertTypeText,
					CurrentPrice:  price.Price,
					Condition:     conditionText,
					TargetValue:   alert.TargetValue,
					TriggerTime:   now.Format("2006-01-02 15:04:05"),
					Change:        price.Change,
					ChangePercent: price.ChangePercent,
				}
				go a.pluginManager.SendNotificationToAll(notifyData)
			}
		}
	}

	return notifications, nil
}

// GetFundAlerts 获取基金提醒
func (a *App) GetFundAlerts(fundCode string) ([]models.FundAlert, error) {
	var alerts []models.FundAlert
	query := data.GetDB().Where("enabled = ?", true)
	if fundCode != "" {
		query = query.Where("fund_code = ?", fundCode)
	}
	err := query.Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// AddFundAlert 添加基金提醒
func (a *App) AddFundAlert(alert models.FundAlert) error {
	alert.Enabled = true
	alert.Triggered = false
	return data.GetDB().Create(&alert).Error
}

// DeleteFundAlert 删除基金提醒
func (a *App) DeleteFundAlert(id uint) error {
	return data.GetDB().Delete(&models.FundAlert{}, id).Error
}

// ToggleFundAlert 切换基金提醒状态
func (a *App) ToggleFundAlert(id uint, enabled bool) error {
	return data.GetDB().Model(&models.FundAlert{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// ResetFundAlert 重置基金提醒
func (a *App) ResetFundAlert(id uint) error {
	return data.GetDB().Model(&models.FundAlert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"triggered":    false,
		"triggered_at": nil,
	}).Error
}

// CheckFundAlerts 检查基金提醒
func (a *App) CheckFundAlerts() ([]models.AlertNotification, error) {
	var alerts []models.FundAlert
	err := data.GetDB().Where("enabled = ? AND triggered = ?", true, false).Find(&alerts).Error
	if err != nil {
		return nil, err
	}
	if len(alerts) == 0 {
		return nil, nil
	}

	a.fundPriceCacheLock.RLock()
	prices := make(map[string]*models.FundPrice, len(a.fundPriceCache))
	for code, price := range a.fundPriceCache {
		prices[code] = price
	}
	a.fundPriceCacheLock.RUnlock()
	if len(prices) == 0 {
		return nil, nil
	}

	var pushConfig *models.Config
	if cfg, err := a.GetConfig(); err == nil && cfg.AlertPushEnabled {
		pushConfig = cfg
	}

	var notifications []models.AlertNotification
	now := time.Now()

	for _, alert := range alerts {
		price, ok := prices[alert.FundCode]
		if !ok || price == nil {
			continue
		}

		currentNav := price.Estimate
		if currentNav == 0 {
			currentNav = price.Nav
		}
		currentChange := price.ChangePercent

		triggered := false
		var message string

		switch alert.AlertType {
		case "nav":
			if alert.Condition == "above" && currentNav >= alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 净值已达到 %.4f（目标：%.4f）", alert.FundName, currentNav, alert.TargetValue)
			} else if alert.Condition == "below" && currentNav <= alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 净值已跌至 %.4f（目标：%.4f）", alert.FundName, currentNav, alert.TargetValue)
			}
		case "change":
			if alert.Condition == "above" && currentChange >= alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 涨幅已达 %.2f%%（目标：%.2f%%）", alert.FundName, currentChange, alert.TargetValue)
			} else if alert.Condition == "below" && currentChange <= -alert.TargetValue {
				triggered = true
				message = fmt.Sprintf("%s 跌幅已达 %.2f%%（目标：-%.2f%%）", alert.FundName, currentChange, alert.TargetValue)
			}
		}

		if !triggered {
			continue
		}

		data.GetDB().Model(&alert).Updates(map[string]interface{}{
			"triggered":        true,
			"triggered_at":     now,
			"triggered_nav":    currentNav,
			"triggered_change": currentChange,
		})

		notification := models.AlertNotification{
			ID:            alert.ID,
			StockCode:     alert.FundCode,
			StockName:     alert.FundName,
			AlertType:     alert.AlertType,
			TargetValue:   alert.TargetValue,
			CurrentPrice:  currentNav,
			CurrentChange: currentChange,
			Message:       message,
			Time:          now.Format("15:04:05"),
			AssetType:     "fund",
		}

		notifications = append(notifications, notification)
		wailsRuntime.EventsEmit(a.ctx, "fund-alert-triggered", notification)

		if pushConfig != nil {
			go a.dispatchAlertPush(pushConfig, notification)
		}

		if a.pluginManager.HasEnabledNotificationPlugins() {
			alertTypeText := "基金净值提醒"
			if alert.AlertType == "change" {
				alertTypeText = "基金涨跌提醒"
			}
			conditionText := "高于"
			if alert.Condition == "below" {
				conditionText = "低于"
			}

			notifyData := &plugin.NotificationData{
				StockCode:     alert.FundCode,
				StockName:     alert.FundName,
				AlertType:     alertTypeText,
				CurrentPrice:  currentNav,
				Condition:     conditionText,
				TargetValue:   alert.TargetValue,
				TriggerTime:   now.Format("2006-01-02 15:04:05"),
				Change:        currentChange,
				ChangePercent: currentChange,
			}
			go a.pluginManager.SendNotificationToAll(notifyData)
		}
	}

	return notifications, nil
}

// TestAlertPush 测试外部推送通道
func (a *App) TestAlertPush(channel string) error {
	cfg, err := a.GetConfig()
	if err != nil {
		return err
	}

	sample := models.AlertNotification{
		ID:            0,
		StockCode:     "sh000001",
		StockName:     "上证指数",
		AlertType:     "price",
		TargetValue:   3200,
		CurrentPrice:  3210,
		CurrentChange: 0.56,
		Message:       "测试提醒：当前价格 3210.00，预设目标 3200.00",
		Time:          time.Now().Format("15:04:05"),
		AssetType:     "stock",
	}

	switch strings.ToLower(channel) {
	case "wecom":
		if strings.TrimSpace(cfg.WecomWebhook) == "" {
			return fmt.Errorf("请先配置企业微信 Webhook")
		}
		return sendWecomWebhook(cfg.WecomWebhook, "Stock AI 提醒测试", formatAlertMarkdown(sample))
	case "dingtalk":
		if strings.TrimSpace(cfg.DingtalkWebhook) == "" {
			return fmt.Errorf("请先配置钉钉 Webhook")
		}
		return sendDingTalkWebhook(cfg.DingtalkWebhook, "Stock AI 提醒测试", formatAlertMarkdown(sample))
	case "email":
		if !cfg.EmailPushEnabled || strings.TrimSpace(cfg.EmailTo) == "" {
			return fmt.Errorf("请先启用并配置邮件推送")
		}
		return sendEmailNotification(cfg, "Stock AI 提醒测试", formatAlertText(sample))
	default:
		return fmt.Errorf("未知的通道: %s", channel)
	}
}

func (a *App) dispatchAlertPush(cfg *models.Config, notification models.AlertNotification) {
	if cfg == nil {
		return
	}
	assetLabel := "股票"
	if notification.AssetType == "fund" {
		assetLabel = "基金"
	}
	title := fmt.Sprintf("%s%s提醒", notification.StockName, assetLabel)
	markdown := formatAlertMarkdown(notification)
	text := formatAlertText(notification)

	if hook := strings.TrimSpace(cfg.WecomWebhook); hook != "" {
		if err := sendWecomWebhook(hook, title, markdown); err != nil {
			log.Printf("[AlertPush] WeCom 推送失败: %v", err)
		}
	}
	if hook := strings.TrimSpace(cfg.DingtalkWebhook); hook != "" {
		if err := sendDingTalkWebhook(hook, title, markdown); err != nil {
			log.Printf("[AlertPush] 钉钉推送失败: %v", err)
		}
	}
	if cfg.EmailPushEnabled && strings.TrimSpace(cfg.EmailTo) != "" {
		if err := sendEmailNotification(cfg, "Stock AI "+title, text); err != nil {
			log.Printf("[AlertPush] 邮件推送失败: %v", err)
		}
	}
}

func formatAlertMarkdown(n models.AlertNotification) string {
	var sb strings.Builder
	priceLabel := "现价"
	if n.AssetType == "fund" {
		priceLabel = "净值"
	}
	sb.WriteString(fmt.Sprintf("**%s (%s)**\n", n.StockName, n.StockCode))
	sb.WriteString(fmt.Sprintf("> %s\n", n.Message))
	sb.WriteString(fmt.Sprintf("> %s：%.4f | 涨跌：%.2f%%\n", priceLabel, n.CurrentPrice, n.CurrentChange))
	sb.WriteString(fmt.Sprintf("> 目标：%.2f | 时间：%s\n", n.TargetValue, n.Time))
	return sb.String()
}

func formatAlertText(n models.AlertNotification) string {
	priceLabel := "现价"
	if n.AssetType == "fund" {
		priceLabel = "净值"
	}
	return fmt.Sprintf(
		"%s (%s)\n%s\n%s: %.4f\n涨跌幅: %.2f%%\n目标值: %.2f\n时间: %s\n",
		n.StockName,
		n.StockCode,
		n.Message,
		priceLabel,
		n.CurrentPrice,
		n.CurrentChange,
		n.TargetValue,
		n.Time,
	)
}

func sendWecomWebhook(webhook string, title string, content string) error {
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": fmt.Sprintf("**%s**\n%s", title, content),
		},
	}
	return postJSON(webhook, payload)
}

func sendDingTalkWebhook(webhook string, title string, content string) error {
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  fmt.Sprintf("### %s\n%s", title, content),
		},
	}
	return postJSON(webhook, payload)
}

func sendEmailNotification(cfg *models.Config, subject string, body string) error {
	recipients := splitRecipients(cfg.EmailTo)
	if len(recipients) == 0 {
		return fmt.Errorf("未配置收件人")
	}

	addr := fmt.Sprintf("%s:%d", cfg.EmailSMTP, cfg.EmailPort)
	host := cfg.EmailSMTP
	auth := smtp.PlainAuth("", cfg.EmailUser, cfg.EmailPassword, host)

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	msg.WriteString(body)

	if cfg.EmailPort == 465 {
		tlsConfig := &tls.Config{ServerName: host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer client.Quit()

		if cfg.EmailUser != "" {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
		if err := client.Mail(cfg.EmailUser); err != nil {
			return err
		}
		for _, rcpt := range recipients {
			if err := client.Rcpt(strings.TrimSpace(rcpt)); err != nil {
				return err
			}
		}
		writer, err := client.Data()
		if err != nil {
			return err
		}
		if _, err := writer.Write(msg.Bytes()); err != nil {
			return err
		}
		return writer.Close()
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host}
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}
	if cfg.EmailUser != "" {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(cfg.EmailUser); err != nil {
		return err
	}
	for _, rcpt := range recipients {
		if err := client.Rcpt(strings.TrimSpace(rcpt)); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg.Bytes()); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func splitRecipients(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n'
	})
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func postJSON(target string, payload interface{}) error {
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(target, "application/json", bytes.NewReader(dataBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("推送失败: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

// ========== 插件管理 ==========

// GetPlugins 获取所有插件
func (a *App) GetPlugins() []plugin.Plugin {
	return a.pluginManager.GetPlugins()
}

// GetPluginsByType 按类型获取插件
func (a *App) GetPluginsByType(pluginType string) []plugin.Plugin {
	return a.pluginManager.GetPluginsByType(plugin.PluginType(pluginType))
}

// GetPlugin 获取单个插件
func (a *App) GetPlugin(id string) (*plugin.Plugin, error) {
	return a.pluginManager.GetPlugin(id)
}

// AddPlugin 添加插件
func (a *App) AddPlugin(pluginData map[string]interface{}) error {
	// 解析插件数据
	id, _ := pluginData["id"].(string)
	name, _ := pluginData["name"].(string)
	pluginType, _ := pluginData["type"].(string)
	description, _ := pluginData["description"].(string)
	enabled, _ := pluginData["enabled"].(bool)
	config := pluginData["config"]

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	p := &plugin.Plugin{
		ID:          id,
		Name:        name,
		Type:        plugin.PluginType(pluginType),
		Description: description,
		Enabled:     enabled,
		Config:      configJSON,
	}

	return a.pluginManager.AddPlugin(p)
}

// UpdatePlugin 更新插件
func (a *App) UpdatePlugin(pluginData map[string]interface{}) error {
	id, _ := pluginData["id"].(string)
	name, _ := pluginData["name"].(string)
	pluginType, _ := pluginData["type"].(string)
	description, _ := pluginData["description"].(string)
	enabled, _ := pluginData["enabled"].(bool)
	config := pluginData["config"]

	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	p := &plugin.Plugin{
		ID:          id,
		Name:        name,
		Type:        plugin.PluginType(pluginType),
		Description: description,
		Enabled:     enabled,
		Config:      configJSON,
	}

	return a.pluginManager.UpdatePlugin(p)
}

// DeletePlugin 删除插件
func (a *App) DeletePlugin(id string) error {
	return a.pluginManager.DeletePlugin(id)
}

// TogglePlugin 启用/禁用插件
func (a *App) TogglePlugin(id string, enabled bool) error {
	return a.pluginManager.TogglePlugin(id, enabled)
}

// GetNotificationTemplates 获取预置通知模板
func (a *App) GetNotificationTemplates() []plugin.NotificationTemplate {
	return a.pluginManager.GetNotificationTemplates()
}

// CreatePluginFromTemplate 从模板创建插件
func (a *App) CreatePluginFromTemplate(templateID string, name string, params map[string]string) error {
	p, err := a.pluginManager.CreatePluginFromTemplate(templateID, name, params)
	if err != nil {
		return err
	}
	return a.pluginManager.AddPlugin(p)
}

// TestNotification 测试通知
func (a *App) TestNotification(pluginID string) error {
	return a.pluginManager.TestNotification(pluginID)
}

// SendNotificationToAll 发送通知到所有启用的通知插件
func (a *App) SendNotificationToAll(data *plugin.NotificationData) []error {
	return a.pluginManager.SendNotificationToAll(data)
}

// HasEnabledNotificationPlugins 检查是否有启用的通知插件
func (a *App) HasEnabledNotificationPlugins() bool {
	return a.pluginManager.HasEnabledNotificationPlugins()
}

// ImportPlugin 导入插件（从JSON字符串）
func (a *App) ImportPlugin(jsonData string) (*plugin.Plugin, error) {
	return a.pluginManager.ImportPlugin(jsonData)
}

// ExportPlugin 导出插件为JSON字符串
func (a *App) ExportPlugin(id string) (string, error) {
	return a.pluginManager.ExportPlugin(id)
}

// GetPluginsDir 获取插件目录路径
func (a *App) GetPluginsDir() string {
	return a.pluginManager.GetPluginsDir()
}

// OpenPluginsDir 打开插件目录
func (a *App) OpenPluginsDir() error {
	dir := a.pluginManager.GetPluginsDir()
	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// 使用系统默认方式打开目录
	wailsRuntime.BrowserOpenURL(a.ctx, "file:///"+strings.ReplaceAll(dir, "\\", "/"))
	return nil
}

// RefreshPlugins 刷新插件列表（扫描并导入新插件文件）
func (a *App) RefreshPlugins() (int, []string) {
	imported, errors := a.pluginManager.ImportAllPluginFiles()
	var errStrings []string
	for _, err := range errors {
		errStrings = append(errStrings, err.Error())
	}
	return imported, errStrings
}

// ========== 数据源插件 ==========

// GetDatasourceTemplates 获取预置数据源模板
func (a *App) GetDatasourceTemplates() []struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Config      plugin.DatasourceConfig `json:"config"`
} {
	return plugin.DatasourceTemplates
}

// TestDatasource 测试数据源插件
func (a *App) TestDatasource(pluginID string, code string) (*plugin.DatasourceResult, error) {
	return a.pluginManager.TestDatasource(pluginID, code)
}

// FetchQuoteFromPlugin 从指定数据源插件获取行情
func (a *App) FetchQuoteFromPlugin(pluginID string, code string) (*plugin.DatasourceResult, error) {
	return a.pluginManager.FetchQuote(pluginID, code)
}

// FetchQuoteFromAllPlugins 从所有启用的数据源插件获取行情
func (a *App) FetchQuoteFromAllPlugins(code string) (*plugin.DatasourceResult, string, error) {
	return a.pluginManager.FetchQuoteFromAll(code)
}

// HasEnabledDatasourcePlugins 检查是否有启用的数据源插件
func (a *App) HasEnabledDatasourcePlugins() bool {
	return a.pluginManager.HasEnabledDatasourcePlugins()
}

// GetEnabledDatasourcePlugins 获取所有启用的数据源插件
func (a *App) GetEnabledDatasourcePlugins() []plugin.Plugin {
	return a.pluginManager.GetEnabledDatasourcePlugins()
}

// ========== AI模型插件 ==========

// GetAITemplates 获取预置AI模型模板
func (a *App) GetAITemplates() []struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Config      plugin.AIConfig `json:"config"`
} {
	return a.pluginManager.GetAITemplates()
}

// CreateAIPluginFromTemplate 从模板创建AI插件
func (a *App) CreateAIPluginFromTemplate(templateID string, name string, apiKey string, baseURL string, model string) error {
	p, err := a.pluginManager.CreateAIPluginFromTemplate(templateID, name, apiKey, baseURL, model)
	if err != nil {
		return err
	}
	return a.pluginManager.AddPlugin(p)
}

// TestAIPlugin 测试AI插件
func (a *App) TestAIPlugin(pluginID string) (string, error) {
	return a.pluginManager.TestAI(pluginID)
}

// AIChatWithPlugin 使用指定AI插件进行对话
func (a *App) AIChatWithPlugin(pluginID string, messages []plugin.AIChatMessage) (string, error) {
	return a.pluginManager.AIChat(pluginID, messages)
}

// AIChatStreamWithPlugin 使用指定AI插件进行流式对话
func (a *App) AIChatStreamWithPlugin(pluginID string, message string) error {
	messages := []plugin.AIChatMessage{
		{Role: "user", Content: message},
	}

	ch, err := a.pluginManager.AIChatStream(pluginID, messages)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "ai-plugin-error", err.Error())
		return err
	}

	go func() {
		for content := range ch {
			wailsRuntime.EventsEmit(a.ctx, "ai-plugin-stream", content)
		}
		wailsRuntime.EventsEmit(a.ctx, "ai-plugin-done", "")
	}()

	return nil
}

// HasEnabledAIPlugins 检查是否有启用的AI插件
func (a *App) HasEnabledAIPlugins() bool {
	return a.pluginManager.HasEnabledAIPlugins()
}

// GetEnabledAIPlugins 获取所有启用的AI插件
func (a *App) GetEnabledAIPlugins() []plugin.Plugin {
	return a.pluginManager.GetEnabledAIPlugins()
}

// ========== 提示词管理 ==========

// GetPromptTypes 获取所有提示词类型
func (a *App) GetPromptTypes() []struct {
	Type        prompt.PromptType `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
} {
	return prompt.GetPromptTypes()
}

// GetPromptsDir 获取提示词目录
func (a *App) GetPromptsDir() string {
	return getPromptsDir()
}

// ListPrompts 列出指定类型的提示词
func (a *App) ListPrompts(promptType string) ([]prompt.PromptInfo, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.List(prompt.PromptType(promptType))
}

// ListAllPrompts 列出所有提示词
func (a *App) ListAllPrompts() (map[string][]prompt.PromptInfo, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}
	result, err := a.promptManager.ListAll()
	if err != nil {
		return nil, err
	}

	// 转换key类型
	converted := make(map[string][]prompt.PromptInfo)
	for k, v := range result {
		converted[string(k)] = v
	}
	return converted, nil
}

// GetPrompt 获取指定提示词
func (a *App) GetPrompt(promptType string, name string) (*prompt.PromptInfo, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.Get(prompt.PromptType(promptType), name)
}

// CreatePrompt 创建提示词
func (a *App) CreatePrompt(promptType string, name string, content string) (*prompt.PromptInfo, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.Create(prompt.PromptType(promptType), name, content)
}

// UpdatePrompt 更新提示词
func (a *App) UpdatePrompt(promptType string, name string, content string) (*prompt.PromptInfo, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.Update(prompt.PromptType(promptType), name, content)
}

// DeletePrompt 删除提示词
func (a *App) DeletePrompt(promptType string, name string) error {
	if a.promptManager == nil {
		return fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.Delete(prompt.PromptType(promptType), name)
}

// RenamePrompt 重命名提示词
func (a *App) RenamePrompt(promptType string, oldName string, newName string) error {
	if a.promptManager == nil {
		return fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.Rename(prompt.PromptType(promptType), oldName, newName)
}

// ImportPrompt 导入提示词
func (a *App) ImportPrompt(promptType string, name string, content string) (*prompt.PromptInfo, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.Import(prompt.PromptType(promptType), name, content)
}

// ExportPrompt 导出提示词
func (a *App) ExportPrompt(promptType string, name string) (string, error) {
	if a.promptManager == nil {
		return "", fmt.Errorf("提示词管理器未初始化")
	}
	return a.promptManager.Export(prompt.PromptType(promptType), name)
}

// OpenPromptsDir 打开提示词目录
func (a *App) OpenPromptsDir() error {
	dir := getPromptsDir()
	wailsRuntime.BrowserOpenURL(a.ctx, "file://"+dir)
	return nil
}

// ========== 提示词执行 ==========

// ExecuteIndicatorPrompt 执行指标提示词分析
func (a *App) ExecuteIndicatorPrompt(promptName string, stockCode string) (*prompt.IndicatorResult, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}

	// 获取提示词
	promptInfo, err := a.promptManager.Get(prompt.PromptTypeIndicator, promptName)
	if err != nil {
		return nil, fmt.Errorf("获取提示词失败: %w", err)
	}

	// 获取股票数据
	stockData, err := a.getStockDataForPrompt(stockCode)
	if err != nil {
		return nil, err
	}

	// 构建提示词
	builtPrompt := prompt.BuildPrompt(promptInfo.Content, stockData)

	// 调用AI
	aiResponse, err := a.callAIForPrompt(builtPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析结果
	result := &prompt.IndicatorResult{
		Signal: prompt.ParseSignal(aiResponse),
		Text:   prompt.TruncateText(aiResponse, 500),
		Raw:    aiResponse,
	}

	return result, nil
}

// ExecuteScreenerPrompt 执行选股提示词
func (a *App) ExecuteScreenerPrompt(promptName string) (*prompt.ScreenerResult, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}

	// 获取提示词
	promptInfo, err := a.promptManager.Get(prompt.PromptTypeScreener, promptName)
	if err != nil {
		return nil, fmt.Errorf("获取提示词失败: %w", err)
	}

	// 获取自选股列表数据
	stocks, err := a.GetStockList()
	if err != nil {
		return nil, fmt.Errorf("获取股票列表失败: %w", err)
	}

	// 获取股票价格
	var codes []string
	for _, s := range stocks {
		codes = append(codes, s.Code)
	}

	prices, err := a.stockAPI.GetStockPrice(codes)
	if err != nil {
		return nil, fmt.Errorf("获取股票价格失败: %w", err)
	}

	// 构建股票数据列表
	var stockDataList []*prompt.StockData
	for _, s := range stocks {
		if price, ok := prices[s.Code]; ok {
			stockDataList = append(stockDataList, &prompt.StockData{
				Code:          s.Code,
				Name:          price.Name,
				Price:         price.Price,
				Change:        price.Change,
				ChangePercent: price.ChangePercent,
				Volume:        float64(price.Volume),
				Amount:        price.Amount,
				High:          price.High,
				Low:           price.Low,
				Open:          price.Open,
				PreClose:      price.PreClose,
			})
		}
	}

	// 构建提示词
	builtPrompt := prompt.BuildPromptWithStockList(promptInfo.Content, stockDataList)

	// 调用AI
	aiResponse, err := a.callAIForPrompt(builtPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析结果
	result := &prompt.ScreenerResult{
		Summary: prompt.TruncateText(aiResponse, 500),
		Raw:     aiResponse,
	}

	return result, nil
}

// ExecuteReviewPrompt 执行复盘提示词
func (a *App) ExecuteReviewPrompt(promptName string) (*prompt.ReviewResult, error) {
	if a.promptManager == nil {
		return nil, fmt.Errorf("提示词管理器未初始化")
	}

	// 获取提示词
	promptInfo, err := a.promptManager.Get(prompt.PromptTypeReview, promptName)
	if err != nil {
		return nil, fmt.Errorf("获取提示词失败: %w", err)
	}

	// 获取持仓数据
	positions, err := a.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	// 获取当前价格
	var codes []string
	for _, p := range positions {
		codes = append(codes, p.StockCode)
	}

	prices, err := a.stockAPI.GetStockPrice(codes)
	if err != nil {
		return nil, fmt.Errorf("获取股票价格失败: %w", err)
	}

	// 构建持仓数据列表
	var positionDataList []*prompt.PositionData
	for _, p := range positions {
		currentPrice := p.CostPrice
		name := p.StockCode
		if price, ok := prices[p.StockCode]; ok {
			currentPrice = price.Price
			name = price.Name
		}
		positionDataList = append(positionDataList, &prompt.PositionData{
			Code:         p.StockCode,
			Name:         name,
			Quantity:     p.Quantity,
			CostPrice:    p.CostPrice,
			CurrentPrice: currentPrice,
		})
	}

	// 构建提示词
	builtPrompt := prompt.BuildPromptWithPortfolio(promptInfo.Content, positionDataList)

	// 调用AI
	aiResponse, err := a.callAIForPrompt(builtPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析结果
	result := &prompt.ReviewResult{
		Summary: prompt.TruncateText(aiResponse, 500),
		Raw:     aiResponse,
	}

	return result, nil
}

// GetActivePersona 获取当前激活的AI人设
func (a *App) GetActivePersona() (string, error) {
	if a.promptManager == nil {
		return "", nil
	}

	// 从配置中获取激活的人设名称
	config, err := a.GetConfig()
	if err != nil {
		return "", nil
	}

	personaName := config.ActivePersona
	if personaName == "" {
		return "", nil
	}

	// 获取人设内容
	promptInfo, err := a.promptManager.Get(prompt.PromptTypePersona, personaName)
	if err != nil {
		return "", nil
	}

	return promptInfo.Content, nil
}

// SetActivePersona 设置激活的AI人设
func (a *App) SetActivePersona(personaName string) error {
	config, err := a.GetConfig()
	if err != nil {
		return err
	}

	config.ActivePersona = personaName
	return a.SaveConfig(*config)
}

// getStockDataForPrompt 获取用于提示词的股票数据
func (a *App) getStockDataForPrompt(stockCode string) (*prompt.StockData, error) {
	stockCode = normalizeStockCode(stockCode)

	stock, err := a.getPriceSnapshot(stockCode)
	if err != nil {
		return nil, fmt.Errorf("获取股票价格失败: %w", err)
	}

	// 获取K线数据
	klines, _ := a.getKLineDataCached(stockCode, 30)

	// 构建股票数据
	stockData := &prompt.StockData{
		Code:          stockCode,
		Name:          stock.Name,
		Price:         stock.Price,
		Change:        stock.Change,
		ChangePercent: stock.ChangePercent,
		Volume:        float64(stock.Volume),
		Amount:        stock.Amount,
		High:          stock.High,
		Low:           stock.Low,
		Open:          stock.Open,
		PreClose:      stock.PreClose,
	}

	// 转换K线数据
	for _, k := range klines {
		stockData.KLines = append(stockData.KLines, prompt.KLineData{
			Date:   k.Date,
			Open:   k.Open,
			Close:  k.Close,
			High:   k.High,
			Low:    k.Low,
			Volume: float64(k.Volume),
		})
	}

	return stockData, nil
}

// callAIForPrompt 调用AI执行提示词
func (a *App) callAIForPrompt(promptText string) (string, error) {
	// 优先使用AI插件
	if a.pluginManager.HasEnabledAIPlugins() {
		messages := []plugin.AIChatMessage{
			{Role: "user", Content: promptText},
		}
		result, _, err := a.pluginManager.AIChatFromAll(messages)
		return result, err
	}

	// 使用内置AI
	if a.aiClient == nil {
		return "", fmt.Errorf("AI未配置")
	}

	messages := []data.ChatMessage{
		{Role: "user", Content: promptText},
	}
	return a.aiClient.Chat(messages)
}
