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
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

var (
	buffClient   = req.C()
	csrfToken    string
	buffNickname string
	getHeaders   = map[string]string{}
	buffCount    int64
	buffItems    []buffItem
)

type buffItem struct {
	assetID    string
	buffPrice  float64
	marketName string
	tradeTime  string
	buffLink   string
}

func buffCookies() {
	fileStr, _ := fileutil.ReadFileToString("cookies_buff.txt")
	if len(fileStr) == 0 {
		fmt.Println("buff_cookies.txt数据为空，请检查")
		waitExit()
	}

	u, _ := url.Parse("https://buff.163.com")
	jar, _ := cookiejar.New(nil)
	var cookies []*http.Cookie

	for _, item := range gjson.Parse(fileStr).Array() {
		domain := item.Get("domain").Str

		if strings.Contains(domain, ".163.com") {
		} else {
			fmt.Println("cookies文件不对，请确保在buff页面上导出cookies")
			waitExit()
		}

		name := item.Get("name").Str
		value := item.Get("value").Str

		cookies = append(cookies, &http.Cookie{
			Domain: domain,
			Name:   name,
			Path:   item.Get("path").Str,
			Value:  value,
		})

		if domain == "buff.163.com" && name == "csrf_token" {
			csrfToken = value
		}
	}

	jar.SetCookies(u, cookies)
	buffClient.SetCookieJar(jar)

	getHeaders = map[string]string{
		"accept":           "application/json, text/javascript, */*; q=0.01",
		"accept-language":  "zh-CN,zh;q=0.9",
		"cache-control":    "no-cache",
		"pragma":           "no-cache",
		"referer":          "https://buff.163.com/market/sell_order/to_deliver?game=csgo",
		"sec-fetch-dest":   "empty",
		"sec-fetch-mode":   "cors",
		"sec-fetch-site":   "same-origin",
		"user-agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36",
		"x-requested-with": "XMLHttpRequest",
	}

}

func checkBuffAccount() {
	accountURL := "https://buff.163.com/account/api/user/info"
	resp, _ := buffClient.R().SetHeaders(getHeaders).Get(accountURL)

	respStr := resp.String()
	buffID := gjson.Get(respStr, "data.id").String()

	if len(buffID) == 0 {
		logger.Error().Str("resp", respStr).Msg("buffID is empty")
		fmt.Printf("获取不到账号ID，或重试执行程序，如果无效，请重新导出buff cookies，错误消息:%s\n", resp.String())
		waitExit()
	}

	buffNickname = gjson.Get(respStr, "data.nickname").Str
	mobile := gjson.Get(resp.String(), "data.mobile").Str
	accountInfo := fmt.Sprintf("buff账号信息 昵称: %s 手机号: %s", buffNickname, mobile)
	fmt.Println(accountInfo)

	if err := auth(buffID); err != nil {
		fmt.Println(err.Error())
		waitExit()
	}
}

func getNodeAttr(node *html.Node, expr string, attr string) string {
	oneNode := htmlquery.FindOne(node, expr)
	if oneNode != nil {
		return htmlquery.SelectAttr(oneNode, attr)
	} else {
		fmt.Printf("请与开发者联系，错误信息 expr:%s attr:%s\n", expr, attr)
		waitExit()
	}

	return ""

}

func checkTime(tradeTime, buffStart, buffEnd time.Time) bool {

	if tradeTime.After(buffStart) && tradeTime.Before(buffEnd) {
		return true
	}

	if tradeTime == buffStart || tradeTime == buffEnd {
		return true
	}

	return false

}

func buffPage(html string) error {
	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		panic(`htmlquery.Parse error`)
	}

	nodes, err := htmlquery.QueryAll(doc, `//tbody[@class="list_tb_csgo"]//tr`)

	if err != nil {
		panic(`not a valid XPath expression.`)
	}

	for _, node := range nodes {
		marketName := getText(node, `//span[@class="textOne"]`)

		href := getNodeAttr(node, `//div[@class="name-cont"]//a`, "href")

		assetID := getNodeAttr(node, `//div[@data-origin="selling-history"]`, "data-assetid")

		priceStr := getNodeAttr(node, `//span[@data-original-currency="CNY"]`, "data-price")

		tradeTimeStr := getText(node, `//td[@class="c_Gray t_Left"]`)

		tradeTime, err := dateparse.ParseLocal(tradeTimeStr)
		if err != nil {
			fmt.Printf("buff交易记录的截止时间: %s 时间解析出错，请检查相关配置\n", viper.GetString("buffEnd"))
			waitExit()
		}

		if checkTime(tradeTime, buffStart, buffEnd) {
			bf := buffItem{
				assetID:    assetID,
				buffPrice:  cast.ToFloat64(priceStr),
				marketName: marketName,
				tradeTime:  tradeTimeStr,
				buffLink:   fmt.Sprintf("https://buff.163.com%s", href),
			}
			buffItems = append(buffItems, bf)
		}

		if tradeTime.Before(buffStart) {
			return fmt.Errorf("交易时间小于初时时间")
		}

	}

	return nil
}

func buffHistory() {
	var pageSize int64 = 100

	var pageNum int64 = 1

	for {
		URL := fmt.Sprintf("https://buff.163.com/market/sell_order/history?game=csgo&state=success&page_num=%d&page_size=%d",
			pageNum, pageSize)
		fmt.Printf("正在获取buff第%d页的历史记录\n", pageNum)

		resp, _ := buffClient.R().Get(URL)
		err := buffPage(resp.String())
		if err != nil {
			break
		}

		time.Sleep(time.Duration(interval) * time.Second)
		pageNum++
	}

}
