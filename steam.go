package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/araddon/dateparse"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/imroc/req/v3"
	"github.com/rapid7/go-get-proxied/proxy"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

var (
	steamClient  = req.C()
	steamHeaders = map[string]string{
		"accept":          "*/*",
		"accept-language": "zh-CN,zh;q=0.9",
		"cache-control":   "no-cache",
		"pragma":          "no-cache",
		"sec-fetch-dest":  "empty",
		"sec-fetch-mode":  "cors",
		"sec-fetch-site":  "same-origin",
		"user-agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36",
		"Origin":          "https://steamcommunity.com",
	}

	secondMap         = map[string]string{}
	thirdMap          = map[string]float64{}
	classInstanceName = map[string]string{}
	assetName         = map[string]string{}
	steamBalance      int64
	steamKey          string
)

func steamCookies() {
	fileStr, _ := fileutil.ReadFileToString("cookies_steam.txt")
	if len(fileStr) == 0 {
		fmt.Println("steam_cookies.txt数据为空，请检查")
		waitExit()
	}

	if strings.Contains(fileStr, "steampowered.com") {
		fmt.Println("cookies导出的域名是steampowered.com，应该从https://steamcommunity.com导出")
		waitExit()
	}

	u, _ := url.Parse("https://steamcommunity.com")
	jar, _ := cookiejar.New(nil)
	var cookies []*http.Cookie
	var sessionID string
	for _, item := range gjson.Parse(fileStr).Array() {
		domain := item.Get("domain").Str
		name := item.Get("name").Str
		value := item.Get("value").Str

		cookies = append(cookies, &http.Cookie{
			Domain: domain,
			Name:   name,
			Path:   item.Get("path").Str,
			Value:  value,
		})

		if name == "sessionid" {
			sessionID = value
		}

	}

	jar.SetCookies(u, cookies)
	steamClient.SetCookieJar(jar)

	if len(sessionID) == 0 {
		fmt.Printf("steam cookies sessionID为空，请确保cookies正确")
		waitExit()
	}

}

func setProxy() {

	var finalProxy string
	switch {
	case httpProxy > 0:
		finalProxy = fmt.Sprintf("http://127.0.0.1:%d", httpProxy)
	case httpProxy == -1:
		p := proxy.NewProvider("").GetProxy("http", "https://www.baidu.com")
		if p != nil {
			fmt.Printf("当前系统代理配置，代理地址:%s 端口号:%d\n", p.Host(), p.Port())
			finalProxy = fmt.Sprintf("http://%s:%d", p.Host(), p.Port())
		}
	}

	if len(finalProxy) > 0 {
		steamClient.SetProxyURL(finalProxy)
	}
}

func getText(doc *html.Node, expr string) string {
	node := htmlquery.FindOne(doc, expr)
	if node != nil {
		return strings.TrimSpace(htmlquery.InnerText(node))
	} else {
		return ""
	}
}

func checkSteamAccount() {
	accountURL := "https://steamcommunity.com/dev/apikey"
	resp, _ := steamClient.R().SetHeaders(getHeaders).Get(accountURL)

	html := resp.String()

	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		panic(`htmlquery.Parse error`)
	}

	nickname := getText(doc, `//span[@id="account_pulldown"]`)
	accountName := getText(doc, `//span[@class="persona online"]`)
	bodyContents := getText(doc, `//div[@id="bodyContents_ex"]/p`)
	if len(strings.Split(bodyContents, ": ")) == 2 {
		steamKey = strings.Split(bodyContents, ": ")[1]
	}

	if len(accountName) == 0 || len(steamKey) == 0 {
		logger.Error().Str("resp", html).Msg("steamID is empty")
		fmt.Println("获取不到账号名或Steam Web API Key，请检查代理或重试执行程序，如果无效，请重新导出steam cookies")
		waitExit()
	}

	accountInfo := fmt.Sprintf("steam账号信息 昵称: %s 账号名: %s Steam Web API Key: %s",
		nickname, accountName, steamKey)
	fmt.Println(accountInfo)
}

func tradeHistory() {
	afterTime := time.Now().Unix()
	tradeID := "0"

	pageNum := 1

	for {

		t, _ := dateparse.ParseLocal(cast.ToString(afterTime))
		if t.Before(steamStart) {
			break
		}

		fmt.Printf("开始获取steam 交易时间:%s 之前的第%d页交易历史数据\n", t.Format(time.DateTime), pageNum)
		URL := fmt.Sprintf("https://api.steampowered.com/IEconService/GetTradeHistory/v1/?max_trades=50&start_after_time=%d&start_after_tradeid=%s&get_descriptions=1&include_total=1&language=english&key=%s",
			afterTime, tradeID, steamKey)
		resp, err := steamClient.R().SetRetryCount(3).Get(URL)

		if err != nil {
			logger.Error().Str("error", err.Error()).Msg("GetTradeHistory error")
			fmt.Printf("steam记录获取异常，请稍后重试或更换快速稳定的代理 错误代码:%s\n", err.Error())
			waitExit()
		}

		totalTrades := gjson.Get(resp.String(), "response.total_trades").Int()
		more := gjson.Get(resp.String(), "response.more").Bool()
		logger.Info().Int("pageNum", pageNum).Int64("totalTrades", totalTrades).
			Bool("more", more).Msg("steam api")
		trades := gjson.Get(resp.String(), "response.trades").Array()

		for _, trade := range trades {
			tradeID = trade.Get("tradeid").String()
			afterTime = trade.Get("time_init").Int()

			for _, given := range trade.Get("assets_given").Array() {
				secondMap[given.Get("new_assetid").Str] = given.Get("assetid").Str
				classID := given.Get("classid").Str
				instanceID := given.Get("instanceid").Str
				value := fmt.Sprintf("%s:%s", classID, instanceID)
				assetName[given.Get("new_assetid").Str] = value
				assetName[given.Get("assetid").Str] = value
			}

		}

		for _, description := range gjson.Get(resp.String(), "response.descriptions").Array() {
			classID := description.Get("classid").Str
			instanceID := description.Get("instanceid").Str
			marketHashName := description.Get("market_hash_name").Str
			key := fmt.Sprintf("%s:%s", classID, instanceID)
			classInstanceName[key] = marketHashName
		}

		if more == false {
			break
		}
		pageNum++
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func steamHistory() {
	fmt.Println("开始获取steam数据")
	tradeHistory()

	var start int64 = 0

	var pageSize int64 = 100

	var totalCount int64 = 0

	pageNum := 1
	for {
		fmt.Printf("开始获取steam第%d页购买历史数据\n", pageNum)
		URL := fmt.Sprintf("https://steamcommunity.com/market/myhistory/render/?query=&start=%d&count=%d&norender=1",
			start, pageSize)

		resp, err := steamClient.R().Get(URL)

		if err != nil {
			logger.Error().Str("error", err.Error()).Msg("steamHistory error")
			fmt.Printf("steam记录获取异常，请稍后重试或更换快速稳定的代理 错误代码:%s\n", err.Error())
			waitExit()
		}

		html := resp.String()

		success := gjson.Get(html, "success").Bool()
		totalCount = gjson.Get(html, "total_count").Int()

		if success {
		} else {
			fmt.Println("steam历史记录状态不正常")
			waitExit()
		}

		skip := steamPage(resp.String())
		if skip {
			break
		}

		if pageSize+start >= totalCount {
			break
		}

		start += pageSize
		pageNum++
		time.Sleep(time.Duration(interval) * time.Second)
	}

	fmt.Printf("steam历史总记录数是%d条 总金额$%.2f\n", totalCount, cast.ToFloat64(steamBalance*1.0/100))

}

func steamPage(html string) bool {
	var total, check int
	for _, purchase := range gjson.Get(html, "purchases").Map() {
		total++
		newID := purchase.Get("asset.new_id").Str
		amount := purchase.Get("asset.amount").Str
		paidAmount := purchase.Get("paid_amount").Int()
		paidFee := purchase.Get("paid_fee").Int()
		steamBalance += paidAmount + paidFee
		thirdMap[newID] = cast.ToFloat64(float64(paidFee+paidAmount) / 100.0)

		timeSold, _ := dateparse.ParseLocal(cast.ToString(purchase.Get("time_sold").Int()))
		logger.Info().Str("amount", amount).Str("newID", newID).
			Int64("paidAmount", paidAmount).Int64("paidFee", paidFee).
			Time("timeSold", timeSold).
			Msg("purchase")

		if timeSold.Before(steamStart) {
			check++
		}

	}

	if total == check {
		return true
	} else {
		return false
	}
}
