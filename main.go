package main

import (
	"fmt"
	"os"
	"time"

	"github.com/araddon/dateparse"
	"github.com/spf13/viper"
	"github.com/zyedidia/generic/multimap"
)

type steamAsset struct {
	AssetID        string
	steamPrice     float64
	marketName     string
	marketHashName string
}

var (
	steamAssetMap = map[string]steamAsset{}
	marketPrice   = map[string]float64{}
	exchangeRate  float64
	multiKeyMap   = multimap.NewMapSlice[string, float64]()
	steamID       string
	nameHash      = map[string]string{}
	userName      string
	interval      int64

	buffStart time.Time

	buffEnd time.Time

	httpProxy int
)

func main() {

	fmt.Println("buff报表软件，作者微信:bufftools 官网:https://www.bufftools.com")

	viper.SetConfigName("config.txt")            // name of config file (without extension)
	viper.SetConfigType("yaml")                  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")                     // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		logger.Error().Str("error", err.Error()).Msg("read config.txt error")
		fmt.Printf("解析配置文件出错: %s\n", err.Error())
		os.Exit(1)
	}

	exchangeRate = viper.GetFloat64("exchangeRate")
	interval = viper.GetInt64("interval")
	httpProxy = viper.GetInt("httpProxy")

	var err error

	buffStart, err = dateparse.ParseLocal(viper.GetString("buffStart"))
	if err != nil {
		fmt.Printf("buff交易记录的开始时间: %s 时间解析出错，请检查相关配置\n", viper.GetString("buffStart"))
		waitExit()
	}

	buffEnd, err = dateparse.ParseLocal(viper.GetString("buffEnd"))
	if err != nil {
		fmt.Printf("buff交易记录的截止时间: %s 时间解析出错，请检查相关配置\n", viper.GetString("buffEnd"))
		waitExit()
	}

	if buffEnd.Before(buffStart) {
		fmt.Printf("buff交易记录的截止时间: %s 小于开始时间: %s 请先修正此问题\n",
			buffEnd.Format("2006-01-02 15-04-05"), buffStart.Format("2006-01-02 15-04-05"))
		waitExit()
	}

	buffCookies()
	checkBuffAccount()

	setProxy()
	steamCookies()
	checkSteamAccount()

	buffHistory()
	steamHistory()

	export()
	waitExit()
}
