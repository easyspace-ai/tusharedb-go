package marketdata

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GetLongTigerList returns eastmoney billboard for trade date (YYYY-MM-DD).
func (c *Client) GetLongTigerList(date string) ([]LongTigerRank, error) {
	if strings.TrimSpace(date) == "" {
		date = todayInChina()
	}

	type row struct {
		SecuCode         string  `json:"SECUCODE"`
		TradeDate        string  `json:"TRADE_DATE"`
		SecurityCode     string  `json:"SECURITY_CODE"`
		SecurityNameAbbr string  `json:"SECURITY_NAME_ABBR"`
		ClosePrice       float64 `json:"CLOSE_PRICE"`
		ChangeRate       float64 `json:"CHANGE_RATE"`
		AccumAmount      float64 `json:"ACCUM_AMOUNT"`
		BillboardBuyAmt  float64 `json:"BILLBOARD_BUY_AMT"`
		BillboardSellAmt float64 `json:"BILLBOARD_SELL_AMT"`
		BillboardNetAmt  float64 `json:"BILLBOARD_NET_AMT"`
		BillboardDealAmt float64 `json:"BILLBOARD_DEAL_AMT"`
		Explanation      string  `json:"EXPLANATION"`
		TurnoverRate     float64 `json:"TURNOVERRATE"`
		FreeMarketCap    float64 `json:"FREE_MARKET_CAP"`
	}
	type envelope struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Result  *struct {
			Data []row `json:"data"`
		} `json:"result"`
	}

	var payload envelope
	resp, err := c.rJSON().
		SetHeader("Referer", "https://data.eastmoney.com/stock/tradedetail.html").
		SetHeader("User-Agent", c.ua()).
		SetQueryParams(map[string]string{
			"sortColumns": "TURNOVERRATE,TRADE_DATE,SECURITY_CODE",
			"sortTypes":   "-1,-1,1",
			"pageSize":    "200",
			"pageNumber":  "1",
			"reportName":  "RPT_DAILYBILLBOARD_DETAILSNEW",
			"columns":     "SECURITY_CODE,SECUCODE,SECURITY_NAME_ABBR,TRADE_DATE,CLOSE_PRICE,CHANGE_RATE,BILLBOARD_NET_AMT,BILLBOARD_BUY_AMT,BILLBOARD_SELL_AMT,BILLBOARD_DEAL_AMT,ACCUM_AMOUNT,TURNOVERRATE,FREE_MARKET_CAP,EXPLANATION",
			"source":      "WEB",
			"client":      "WEB",
			"filter":      fmt.Sprintf("(TRADE_DATE<='%s')(TRADE_DATE>='%s')", date, date),
		}).
		SetResult(&payload).
		Get("https://datacenter-web.eastmoney.com/api/data/v1/get")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("eastmoney long tiger: %s", resp.Status())
	}
	// 9201 = 返回数据为空（非交易日、当日尚未更新、或未来日期）；前端会按日期回退重试。
	if payload.Code != 0 {
		if payload.Code == 9201 {
			return []LongTigerRank{}, nil
		}
		return nil, fmt.Errorf("eastmoney long tiger: %s (code %d)", strings.TrimSpace(payload.Message), payload.Code)
	}
	if payload.Result == nil || len(payload.Result.Data) == 0 {
		return []LongTigerRank{}, nil
	}

	items := make([]LongTigerRank, 0, len(payload.Result.Data))
	for _, item := range payload.Result.Data {
		items = append(items, LongTigerRank{
			TradeDate:        item.TradeDate,
			SecurityCode:     item.SecurityCode,
			SecuCode:         item.SecuCode,
			SecurityNameAbbr: item.SecurityNameAbbr,
			ClosePrice:       item.ClosePrice,
			ChangeRate:       item.ChangeRate,
			AccumAmount:      item.AccumAmount,
			BillboardBuyAmt:  item.BillboardBuyAmt,
			BillboardSellAmt: item.BillboardSellAmt,
			BillboardNetAmt:  item.BillboardNetAmt,
			BillboardDealAmt: item.BillboardDealAmt,
			Explanation:      item.Explanation,
			TurnoverRate:     item.TurnoverRate,
			FreeMarketCap:    item.FreeMarketCap,
		})
	}
	return items, nil
}

// GetHotStocks fetches xueqiu hot stock list.
func (c *Client) GetHotStocks(source string) ([]HotStock, error) {
	if strings.TrimSpace(source) == "" {
		source = "xueqiu"
	}
	_ = source

	type hotStockResp struct {
		Data struct {
			Items []struct {
				Code       string  `json:"code"`
				Name       string  `json:"name"`
				Value      float64 `json:"value"`
				Increment  int     `json:"increment"`
				RankChange int     `json:"rank_change"`
				Percent    float64 `json:"percent"`
				Current    float64 `json:"current"`
				Chg        float64 `json:"chg"`
				Exchange   string  `json:"exchange"`
			} `json:"items"`
		} `json:"data"`
	}

	var payload hotStockResp
	resp, err := c.rJSON().
		SetHeader("Referer", "https://xueqiu.com/hq#exchange=CN&firstName=1&secondName=1_0").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get("https://stock.xueqiu.com/v5/stock/hot_stock/list.json?page=1&size=20&_type=10&type=10")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("xueqiu hot stock: %s", resp.Status())
	}

	items := make([]HotStock, 0, len(payload.Data.Items))
	for _, item := range payload.Data.Items {
		items = append(items, HotStock{
			Code:       item.Code,
			Name:       item.Name,
			Value:      item.Value,
			Increment:  item.Increment,
			RankChange: item.RankChange,
			Percent:    item.Percent,
			Current:    item.Current,
			Chg:        item.Chg,
			Exchange:   item.Exchange,
		})
	}
	return items, nil
}

// GetHotEvents fetches xueqiu hot events.
func (c *Client) GetHotEvents() ([]HotEvent, error) {
	type hotEventResp struct {
		List []struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Content     string `json:"content"`
			Tag         string `json:"tag"`
			Pic         string `json:"pic"`
			Hot         int    `json:"hot"`
			StatusCount int    `json:"status_count"`
		} `json:"list"`
	}

	var payload hotEventResp
	resp, err := c.rJSON().
		SetHeader("Referer", "https://xueqiu.com/").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get("https://xueqiu.com/hot_event/list.json?count=20")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("xueqiu hot event: %s", resp.Status())
	}

	items := make([]HotEvent, 0, len(payload.List))
	for _, item := range payload.List {
		items = append(items, HotEvent{
			ID:          item.ID,
			Title:       item.Title,
			Content:     item.Content,
			Tag:         item.Tag,
			Pic:         item.Pic,
			Hot:         item.Hot,
			StatusCount: item.StatusCount,
		})
	}
	return items, nil
}

// GetHotTopics fetches eastmoney guba hot topics.
func (c *Client) GetHotTopics() ([]HotTopic, error) {
	var payload struct {
		Re []struct {
			HTID        string `json:"htid"`
			Nickname    string `json:"nickname"`
			Desc        string `json:"desc"`
			ClickNumber int    `json:"clickNumber"`
			PostNumber  int    `json:"postNumber"`
		} `json:"re"`
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://gubatopic.eastmoney.com/").
		SetHeader("User-Agent", c.ua()).
		SetFormData(map[string]string{
			"param": "ps=10&p=1&type=0",
			"path":  "newtopic/api/Topic/HomePageListRead",
			"env":   "2",
		}).
		SetResult(&payload).
		Post("https://gubatopic.eastmoney.com/interface/GetData.aspx?path=newtopic/api/Topic/HomePageListRead")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("eastmoney hot topic: %s", resp.Status())
	}

	items := make([]HotTopic, 0, len(payload.Re))
	for idx, item := range payload.Re {
		items = append(items, HotTopic{
			ID:         idx + 1,
			Title:      item.Nickname,
			Content:    item.Desc,
			Hot:        item.ClickNumber,
			StockCount: item.PostNumber,
		})
	}
	return items, nil
}

// GetNews24h returns Cailian telegraph-style list (client-side paginated).
func (c *Client) GetNews24h(page, pageSize int) ([]MarketNews, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	var payload struct {
		Data struct {
			RollData []struct {
				Title       string `json:"title"`
				Content     string `json:"content"`
				Ctime       any    `json:"ctime"`
				Url         string `json:"url"`
				Tag         string `json:"tag"`
				StockList   any    `json:"stock_list"`
				Source      string `json:"source"`
				Market      string `json:"market"`
				Description string `json:"description"`
			} `json:"roll_data"`
		} `json:"data"`
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://www.cls.cn/telegraph").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get("https://www.cls.cn/nodeapi/telegraphList?app=CailianpressWeb&os=web&sv=7.7.5")
	if err != nil {
		return nil, 0, err
	}
	if resp.IsError() {
		return nil, 0, fmt.Errorf("cls news24h: %s", resp.Status())
	}

	all := make([]MarketNews, 0, len(payload.Data.RollData))
	for idx, item := range payload.Data.RollData {
		all = append(all, MarketNews{
			ID:          uint(idx + 1),
			Title:       firstNonEmpty(item.Title, item.Tag, "市场快讯"),
			Content:     firstNonEmpty(item.Content, item.Description),
			Source:      firstNonEmpty(item.Source, "财联社"),
			Url:         item.Url,
			PublishTime: parseUnixOrNow(fmt.Sprintf("%v", item.Ctime)),
			StockCodes:  joinAnyList(item.StockList),
			Tags:        item.Market,
		})
	}

	return paginate(all, page, pageSize), int64(len(all)), nil
}

// GetSinaNews returns sina finance live feed (client-side paginated).
func (c *Client) GetSinaNews(page, pageSize int) ([]MarketNews, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://finance.sina.com.cn").
		SetHeader("User-Agent", c.ua()).
		Get("https://zhibo.sina.com.cn/api/zhibo/feed?callback=callback&page=1&page_size=20&zhibo_id=152&tag_id=0&dire=f&dpc=1&pagesize=20&id=4161089&type=0&_=" + strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		return nil, 0, err
	}
	if resp.IsError() {
		return nil, 0, fmt.Errorf("sina news: %s", resp.Status())
	}

	body := string(resp.Body())
	body = strings.TrimPrefix(body, "try{callback(")
	body = strings.TrimSuffix(body, ");}catch(e){};")

	var payload struct {
		Result struct {
			Data struct {
				Feed struct {
					List []struct {
						RichText   string `json:"rich_text"`
						CreateTime string `json:"create_time"`
						Tag        []struct {
							Name string `json:"name"`
						} `json:"tag"`
					} `json:"list"`
				} `json:"feed"`
			} `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil, 0, err
	}

	all := make([]MarketNews, 0, len(payload.Result.Data.Feed.List))
	for idx, item := range payload.Result.Data.Feed.List {
		tags := make([]string, 0, len(item.Tag))
		for _, tag := range item.Tag {
			if strings.TrimSpace(tag.Name) != "" {
				tags = append(tags, tag.Name)
			}
		}
		title := ""
		if strings.Contains(item.RichText, "【") && strings.Contains(item.RichText, "】") {
			parts := strings.SplitN(strings.SplitN(item.RichText, "】", 2)[0], "【", 2)
			if len(parts) == 2 {
				title = parts[1]
			}
		}
		publishTime, _ := time.ParseInLocation("2006-01-02 15:04:05", item.CreateTime, time.Local)
		if publishTime.IsZero() {
			publishTime = time.Now()
		}
		all = append(all, MarketNews{
			ID:          uint(idx + 1),
			Title:       firstNonEmpty(title, "新浪财经"),
			Content:     item.RichText,
			Source:      "新浪财经",
			Url:         "",
			PublishTime: publishTime,
			StockCodes:  "",
			Tags:        strings.Join(tags, ","),
		})
	}

	return paginate(all, page, pageSize), int64(len(all)), nil
}

// GetStockNews is not implemented upstream (legacy stub).
func (c *Client) GetStockNews(_ string, _ int, _ int) ([]MarketNews, int64, error) {
	return []MarketNews{}, 0, nil
}

// GetStockResearchReport fetches eastmoney stock research list.
func (c *Client) GetStockResearchReport(code string, page, pageSize int) ([]ResearchReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	code = normalizeStockCode(code)

	type requestBody struct {
		BeginTime    string `json:"beginTime"`
		EndTime      string `json:"endTime"`
		IndustryCode string `json:"industryCode"`
		Code         string `json:"code"`
		PageSize     int    `json:"pageSize"`
		PageNo       int    `json:"pageNo"`
		P            int    `json:"p"`
		PageNum      int    `json:"pageNum"`
		PageNumber   int    `json:"pageNumber"`
	}
	type reportResp struct {
		Data []struct {
			Title       string `json:"title"`
			StockCode   string `json:"stockCode"`
			StockName   string `json:"stockName"`
			Researcher  string `json:"researcher"`
			OrgSName    string `json:"orgSName"`
			PublishDate string `json:"publishDate"`
			InfoCode    string `json:"infoCode"`
			EmRating    string `json:"emRatingName"`
			Industry    string `json:"indvInduName"`
		} `json:"data"`
		TotalHits int64 `json:"TotalHits"`
	}

	var payload reportResp
	resp, err := c.rJSON().
		SetHeader("Origin", "https://data.eastmoney.com").
		SetHeader("Referer", "https://data.eastmoney.com/report/stock.jshtml").
		SetHeader("User-Agent", c.ua()).
		SetBody(requestBody{
			Code:         code,
			IndustryCode: "*",
			BeginTime:    time.Now().Add(-365 * 24 * time.Hour).Format("2006-01-02"),
			EndTime:      time.Now().Format("2006-01-02"),
			PageNo:       page,
			PageSize:     pageSize,
			P:            page,
			PageNum:      page,
			PageNumber:   page,
		}).
		SetResult(&payload).
		Post("https://reportapi.eastmoney.com/report/list2")
	if err != nil {
		return nil, 0, err
	}
	if resp.IsError() {
		return nil, 0, fmt.Errorf("eastmoney stock report: %s", resp.Status())
	}

	items := make([]ResearchReport, 0, len(payload.Data))
	for idx, item := range payload.Data {
		items = append(items, ResearchReport{
			ID:          uint(idx + 1),
			Title:       item.Title,
			Content:     firstNonEmpty(item.EmRating, item.Industry),
			StockCode:   item.StockCode,
			StockName:   item.StockName,
			Author:      item.Researcher,
			OrgName:     item.OrgSName,
			PublishDate: item.PublishDate,
			ReportType:  "stock",
			Url:         buildEastmoneyReportPDF(item.InfoCode),
		})
	}
	return items, payload.TotalHits, nil
}

// GetIndustryResearchReport fetches eastmoney industry research list.
func (c *Client) GetIndustryResearchReport(industry string, page, pageSize int) ([]ResearchReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	type reportResp struct {
		Data []struct {
			Title        string `json:"title"`
			IndustryName string `json:"industryName"`
			Researcher   string `json:"researcher"`
			OrgSName     string `json:"orgSName"`
			PublishDate  string `json:"publishDate"`
			InfoCode     string `json:"infoCode"`
			EmRatingName string `json:"emRatingName"`
		} `json:"data"`
		TotalHits int64 `json:"TotalHits"`
	}

	var payload reportResp
	resp, err := c.rJSON().
		SetHeader("Origin", "https://data.eastmoney.com").
		SetHeader("Referer", "https://data.eastmoney.com/report/industry.jshtml").
		SetHeader("User-Agent", c.ua()).
		SetQueryParams(map[string]string{
			"industry":     "*",
			"industryCode": strings.TrimSpace(industry),
			"beginTime":    time.Now().Add(-365 * 24 * time.Hour).Format("2006-01-02"),
			"endTime":      time.Now().Format("2006-01-02"),
			"pageNo":       strconv.Itoa(page),
			"pageSize":     strconv.Itoa(pageSize),
			"p":            strconv.Itoa(page),
			"pageNum":      strconv.Itoa(page),
			"pageNumber":   strconv.Itoa(page),
			"qType":        "1",
		}).
		SetResult(&payload).
		Get("https://reportapi.eastmoney.com/report/list")
	if err != nil {
		return nil, 0, err
	}
	if resp.IsError() {
		return nil, 0, fmt.Errorf("eastmoney industry report: %s", resp.Status())
	}

	items := make([]ResearchReport, 0, len(payload.Data))
	for idx, item := range payload.Data {
		items = append(items, ResearchReport{
			ID:          uint(idx + 1),
			Title:       item.Title,
			Content:     item.EmRatingName,
			StockCode:   strings.TrimSpace(industry),
			StockName:   item.IndustryName,
			Author:      item.Researcher,
			OrgName:     item.OrgSName,
			PublishDate: item.PublishDate,
			ReportType:  "industry",
			Url:         buildEastmoneyReportPDF(item.InfoCode),
		})
	}
	return items, payload.TotalHits, nil
}

// GetStockNotice fetches eastmoney stock announcements.
func (c *Client) GetStockNotice(code string, page, pageSize int) ([]StockNotice, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var stockCodes []string
	if strings.TrimSpace(code) != "" {
		for _, stockCode := range strings.Split(code, ",") {
			normalized := normalizeStockCode(stockCode)
			if normalized != "" {
				stockCodes = append(stockCodes, normalized)
			}
		}
	}

	url := fmt.Sprintf(
		"https://np-anotice-stock.eastmoney.com/api/security/ann?page_size=%d&page_index=%d&ann_type=SHA%%2CCYB%%2CSZA%%2CBJA%%2CINV&client_source=web&f_node=0&stock_list=%s",
		pageSize,
		page,
		strings.Join(stockCodes, ","),
	)

	resp, err := c.rJSON().
		SetHeader("Host", "np-anotice-stock.eastmoney.com").
		SetHeader("Referer", "https://data.eastmoney.com/notices/hsa/5.html").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0").
		Get(url)
	if err != nil {
		return nil, 0, err
	}
	if resp.IsError() {
		return nil, 0, fmt.Errorf("eastmoney notice: %s", resp.Status())
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body(), &payload); err != nil {
		return nil, 0, err
	}

	dataMap, ok := payload["data"].(map[string]any)
	if !ok || dataMap == nil {
		return []StockNotice{}, 0, nil
	}

	listRaw, ok := dataMap["list"].([]any)
	if !ok || listRaw == nil {
		return []StockNotice{}, 0, nil
	}

	totalHits := int64(parseFloat(fmt.Sprintf("%v", dataMap["total_hits"])))
	items := make([]StockNotice, 0, len(listRaw))
	for idx, raw := range listRaw {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		stockCode := ""
		stockName := ""
		if codes, ok := item["codes"].([]any); ok && len(codes) > 0 {
			if firstCode, ok := codes[0].(map[string]any); ok {
				stockCode = fmt.Sprintf("%v", firstCode["stock_code"])
				stockName = fmt.Sprintf("%v", firstCode["short_name"])
			}
		}

		noticeType := ""
		if columns, ok := item["columns"].([]any); ok && len(columns) > 0 {
			if firstColumn, ok := columns[0].(map[string]any); ok {
				noticeType = fmt.Sprintf("%v", firstColumn["column_name"])
			}
		}

		title := fmt.Sprintf("%v", item["title"])
		artCode := fmt.Sprintf("%v", item["art_code"])
		noticeDate := fmt.Sprintf("%v", item["notice_date"])
		displayTime := fmt.Sprintf("%v", item["display_time"])
		items = append(items, StockNotice{
			ID:          uint(idx + 1),
			Title:       title,
			Content:     noticeType,
			StockCode:   stockCode,
			StockName:   stockName,
			NoticeType:  noticeType,
			PublishDate: noticeDate,
			UpdateTime:  displayTime,
			Url:         buildEastmoneyNoticePDF(artCode),
		})
	}
	return items, totalHits, nil
}

// GetIndustryRank returns QQ finance sector rank.
func (c *Client) GetIndustryRank(sort string, count int) ([]IndustryRank, error) {
	if strings.TrimSpace(sort) == "" {
		sort = "0"
	}
	if count <= 0 {
		count = 150
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://stockapp.finance.qq.com/").
		SetHeader("User-Agent", c.ua()).
		Get(fmt.Sprintf("https://proxy.finance.qq.com/ifzqgtimg/appstock/app/mktHs/rank?l=%d&p=1&t=01/averatio&ordertype=&o=%s", count, sort))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("qq industry rank: %s", resp.Status())
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body(), &payload); err != nil {
		return nil, err
	}

	pageData, ok := payload["data"].([]any)
	if !ok || pageData == nil {
		dataMap, ok := payload["data"].(map[string]any)
		if !ok || dataMap == nil {
			return []IndustryRank{}, nil
		}
		if nested, ok := dataMap["page_data"].([]any); ok {
			pageData = nested
		} else {
			return []IndustryRank{}, nil
		}
	}

	items := make([]IndustryRank, 0, len(pageData))
	for _, raw := range pageData {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		items = append(items, IndustryRank{
			IndustryName:  fmt.Sprintf("%v", item["bd_name"]),
			IndustryCode:  fmt.Sprintf("%v", item["bd_code"]),
			ChangePct:     parseFloat(fmt.Sprintf("%v", item["bd_zdf"])),
			ChangePct5d:   parseFloat(fmt.Sprintf("%v", item["bd_zdf5"])),
			ChangePct20d:  parseFloat(fmt.Sprintf("%v", item["bd_zdf20"])),
			LeadStock:     fmt.Sprintf("%v", item["nzg_name"]),
			LeadStockCode: fmt.Sprintf("%v", item["nzg_code"]),
			LeadChange:    parseFloat(fmt.Sprintf("%v", item["nzg_zdf"])),
			LeadPrice:     parseFloat(fmt.Sprintf("%v", item["nzg_zxj"])),
		})
	}
	return items, nil
}

// GetIndustryMoneyRank returns sina sector money flow rank.
func (c *Client) GetIndustryMoneyRank(fenlei, sort string) ([]IndustryMoneyRank, error) {
	if strings.TrimSpace(fenlei) == "" {
		fenlei = "0"
	}
	if strings.TrimSpace(sort) == "" {
		sort = "netamount"
	}

	var payload []struct {
		Name           string `json:"name"`
		AvgChangeRatio string `json:"avg_changeratio"`
		InAmount       string `json:"inamount"`
		OutAmount      string `json:"outamount"`
		NetAmount      string `json:"netamount"`
		RatioAmount    string `json:"ratioamount"`
		TSName         string `json:"ts_name"`
		TSSymbol       string `json:"ts_symbol"`
		TSTrade        string `json:"ts_trade"`
		TSChangeRatio  string `json:"ts_changeratio"`
		TSRatioAmount  string `json:"ts_ratioamount"`
	}

	resp, err := c.rJSON().
		SetHeader("Host", "vip.stock.finance.sina.com.cn").
		SetHeader("Referer", "https://finance.sina.com.cn").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get(fmt.Sprintf("https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/MoneyFlow.ssl_bkzj_bk?page=1&num=20&sort=%s&asc=0&fenlei=%s", sort, fenlei))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("sina industry money: %s", resp.Status())
	}

	items := make([]IndustryMoneyRank, 0, len(payload))
	for _, item := range payload {
		items = append(items, IndustryMoneyRank{
			IndustryName:  item.Name,
			ChangePct:     parseFloat(item.AvgChangeRatio) * 100,
			Inflow:        parseFloat(item.InAmount),
			Outflow:       parseFloat(item.OutAmount),
			NetInflow:     parseFloat(item.NetAmount),
			NetRatio:      parseFloat(item.RatioAmount) * 100,
			LeadStock:     item.TSName,
			LeadStockCode: item.TSSymbol,
			LeadChange:    parseFloat(item.TSChangeRatio) * 100,
			LeadPrice:     parseFloat(item.TSTrade),
			LeadNetRatio:  parseFloat(item.TSRatioAmount) * 100,
		})
	}
	return items, nil
}

// GetStockMoneyRank returns sina stock money flow rank.
func (c *Client) GetStockMoneyRank(sort string) ([]StockMoneyRank, error) {
	if strings.TrimSpace(sort) == "" {
		sort = "netamount"
	}

	var payload []struct {
		Code     string `json:"symbol"`
		Name     string `json:"name"`
		Trade    string `json:"trade"`
		Change   string `json:"changeratio"`
		Turnover string `json:"turnover"`
		Amount   string `json:"amount"`
		OutAmt   string `json:"outamount"`
		InAmt    string `json:"inamount"`
		NetAmt   string `json:"netamount"`
		NetRatio string `json:"ratioamount"`
		R0Out    string `json:"r0_out"`
		R0In     string `json:"r0_in"`
		R0Net    string `json:"r0_net"`
		R0Ratio  string `json:"r0_ratio"`
		R3Out    string `json:"r3_out"`
		R3In     string `json:"r3_in"`
		R3Net    string `json:"r3_net"`
		R3Ratio  string `json:"r3_ratio"`
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://finance.sina.com.cn").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get(fmt.Sprintf("https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/MoneyFlow.ssl_bkzj_ssggzj?page=1&num=20&sort=%s&asc=0&bankuai=&shichang=", sort))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("sina stock money rank: %s", resp.Status())
	}

	items := make([]StockMoneyRank, 0, len(payload))
	for _, item := range payload {
		items = append(items, StockMoneyRank{
			Code:         item.Code,
			Name:         item.Name,
			Price:        parseFloat(item.Trade),
			ChangePct:    parseFloat(item.Change) * 100,
			TurnoverRate: parseFloat(item.Turnover),
			Amount:       parseFloat(item.Amount),
			OutAmount:    parseFloat(item.OutAmt),
			InAmount:     parseFloat(item.InAmt),
			NetAmount:    parseFloat(item.NetAmt),
			NetRatio:     parseFloat(item.NetRatio) * 100,
			R0Out:        parseFloat(item.R0Out),
			R0In:         parseFloat(item.R0In),
			R0Net:        parseFloat(item.R0Net),
			R0Ratio:      parseFloat(item.R0Ratio) * 100,
			R3Out:        parseFloat(item.R3Out),
			R3In:         parseFloat(item.R3In),
			R3Net:        parseFloat(item.R3Net),
			R3Ratio:      parseFloat(item.R3Ratio) * 100,
		})
	}
	return items, nil
}

// GetStockMoneyTrend returns sina per-stock money flow history.
func (c *Client) GetStockMoneyTrend(code string) ([]MoneyFlowInfo, error) {
	code = normalizeStockCode(code)
	var payload []struct {
		OpenDate  string `json:"opendate"`
		NetAmount string `json:"netamount"`
		NetRatio  string `json:"ratioamount"`
		R0Net     string `json:"r0_net"`
		R1Net     string `json:"r1_net"`
		R2Net     string `json:"r2_net"`
		R3Net     string `json:"r3_net"`
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://finance.sina.com.cn").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get(fmt.Sprintf("https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/MoneyFlow.ssl_qsfx_zjlrqs?page=1&num=30&sort=opendate&asc=0&daima=%s", code))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("sina stock money trend: %s", resp.Status())
	}

	items := make([]MoneyFlowInfo, 0, len(payload))
	for _, item := range payload {
		items = append(items, MoneyFlowInfo{
			Date:                item.OpenDate,
			MainNetInflow:       parseFloat(item.NetAmount),
			MainNetRatio:        parseFloat(item.NetRatio) * 100,
			SuperLargeNetInflow: parseFloat(item.R0Net),
			LargeNetInflow:      parseFloat(item.R1Net),
			MediumNetInflow:     parseFloat(item.R2Net),
			SmallNetInflow:      parseFloat(item.R3Net),
		})
	}
	return items, nil
}

// GetGlobalIndexes returns QQ finance global index board.
func (c *Client) GetGlobalIndexes() ([]GlobalIndex, error) {
	var payload struct {
		Data map[string][]struct {
			Code  string `json:"code"`
			Name  string `json:"name"`
			Zxj   string `json:"zxj"`
			Zde   string `json:"zde"`
			Zdf   string `json:"zdf"`
			Time  string `json:"time"`
			State string `json:"state"`
		} `json:"data"`
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://stockapp.finance.qq.com/mstats").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get("https://proxy.finance.qq.com/ifzqgtimg/appstock/app/rank/indexRankDetail2")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("qq global indexes: %s", resp.Status())
	}

	orderedKeys := []string{"common", "america", "asia", "europe", "other"}
	items := make([]GlobalIndex, 0)
	for _, key := range orderedKeys {
		for _, item := range payload.Data[key] {
			items = append(items, GlobalIndex{
				Name:       item.Name,
				Code:       item.Code,
				Price:      parseFloat(item.Zxj),
				Change:     parseFloat(item.Zde),
				ChangePct:  parseFloat(strings.TrimSuffix(item.Zdf, "%")),
				UpdateTime: firstNonEmpty(item.Time, item.State),
			})
		}
	}
	return items, nil
}

// GetInvestCalendar returns jiuyangongshe timeline list for month YYYY-MM.
func (c *Client) GetInvestCalendar(startDate, endDate string) ([]InvestCalendarItem, error) {
	if strings.TrimSpace(startDate) == "" {
		startDate = time.Now().Format("2006-01")
	}
	if len(startDate) >= 7 {
		startDate = startDate[:7]
	}

	var payload struct {
		Data []struct {
			Date  string `json:"date"`
			Title string `json:"title"`
			Desc  string `json:"desc"`
			Type  string `json:"type"`
		} `json:"data"`
	}

	resp, err := c.rJSON().
		SetHeader("Origin", "https://www.jiuyangongshe.com").
		SetHeader("Referer", "https://www.jiuyangongshe.com/").
		SetHeader("User-Agent", c.ua()).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"date":  startDate,
			"grade": "0",
		}).
		SetResult(&payload).
		Post("https://app.jiuyangongshe.com/jystock-app/api/v1/timeline/list")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("jyg invest calendar: %s", resp.Status())
	}

	items := make([]InvestCalendarItem, 0, len(payload.Data))
	for _, item := range payload.Data {
		items = append(items, InvestCalendarItem{
			Date:    item.Date,
			Title:   item.Title,
			Content: item.Desc,
			Type:    item.Type,
		})
	}
	return items, nil
}

// GetCLSCalendar returns CLS web calendar list.
func (c *Client) GetCLSCalendar(startDate, endDate string) ([]InvestCalendarItem, error) {
	_ = startDate
	_ = endDate

	var payload struct {
		Data []struct {
			Date      string `json:"date"`
			Title     string `json:"title"`
			Brief     string `json:"brief"`
			EventType string `json:"event_type"`
		} `json:"data"`
	}

	resp, err := c.rJSON().
		SetHeader("Referer", "https://www.cls.cn/").
		SetHeader("User-Agent", c.ua()).
		SetResult(&payload).
		Get("https://www.cls.cn/api/calendar/web/list?app=CailianpressWeb&flag=0&os=web&sv=8.4.6&type=0&sign=4b839750dc2f6b803d1c8ca00d2b40be")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("cls calendar: %s", resp.Status())
	}

	items := make([]InvestCalendarItem, 0, len(payload.Data))
	for _, item := range payload.Data {
		items = append(items, InvestCalendarItem{
			Date:    item.Date,
			Title:   item.Title,
			Content: item.Brief,
			Type:    item.EventType,
		})
	}
	return items, nil
}
