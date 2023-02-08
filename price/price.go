package main

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

// https://www.zhihu.com/question/37273716

func main() {

	content, _ := fileutil.ReadFileToString("prices_v6.json")

	for marketHashName, v := range gjson.Parse(content).Map() {
		priceMap := map[string]float64{}
		buff163 := v.Get("buff163.starting_at.price").Float()
		steamPrice := v.Get("steam.last_24h").Float()
		if steamPrice == 0 || strings.HasPrefix(marketHashName, "Sticker") {
			continue
		}
		for k2, v2 := range v.Map() {
			if strings.HasPrefix(k2, "buff163") || strings.HasPrefix(k2, "steam") {
				continue
			}
			for k3, v3 := range v2.Map() {
				key := fmt.Sprintf("%s.%s", k2, k3)
				priceMap[key] = cast.ToFloat64(v3.String())
			}
		}

		for marketName, marketPrice := range priceMap {
			if marketPrice > 0 && marketPrice <= buff163*0.84 {
				fmt.Println(marketHashName, marketName, marketPrice, buff163)
			}
		}
	}

}
