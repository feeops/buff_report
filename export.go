package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

func export() {
	fileName := fmt.Sprintf("%s %s.xlsx", buffNickname, time.Now().Format("2006-01-02 15-04-05"))
	fmt.Printf("开始导出利润报表，文件名为%s\n", fileName)
	// 创建文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	titleSlice := []string{
		"名称",
		"英文名",
		"steam购入价格(美元)",
		"卡价",
		"steam购入价格(人民币)",
		"buff销售价格(人民币)",
		"纯利润(已扣除手续费)",
		"纯利润率(已扣除手续费)",
		"成交时间",
		"buff链接",
		"备注",
	}
	_ = f.SetSheetRow("Sheet1", "A1", &titleSlice)

	fmt.Printf("buff指定时间内历史总记录数是%d条,无法匹配上的会自动跳过\n", len(buffItems))

	for k, item := range buffItems {
		var ratio, steamRMB, profit float64
		var remark string
		buffPrice := item.buffPrice
		assetID := item.assetID
		secondAssetID := secondMap[assetID]
		steamPrice := thirdMap[secondAssetID]
		marketHashName := classInstanceName[assetName[item.assetID]]

		if len(marketHashName) == 0 || steamPrice == 0 {
			logger.Error().Str("assetID", assetID).
				Str("secondAssetID", secondAssetID).Float64("buffPrice", buffPrice).
				Msg("empty buff item")
			remark = "获取不到交易记录，饰品购买渠道不是来自于steam市场，请自行处理"
		} else {
			ratio, _ = decimal.
				NewFromFloat(
					(buffPrice*(1-0.025)*(1-0.01) - steamPrice*exchangeRate) * 100 /
						(steamPrice * exchangeRate)).Round(2).Float64()
			steamRMB, _ = decimal.NewFromFloat(exchangeRate * steamPrice).Round(2).Float64()
			profit, _ = decimal.NewFromFloat(
				buffPrice*(1-0.025)*(1-0.01) - steamPrice*exchangeRate).Round(2).Float64()
		}

		data := []interface{}{
			item.marketName,
			marketHashName,
			steamPrice,
			exchangeRate,
			steamRMB,
			buffPrice,
			profit,
			ratio,
			item.tradeTime,
			item.buffLink,
			remark,
		}

		f.SetSheetRow("sheet1", fmt.Sprint("A", k+2), &data)
	}
	// 写入数据
	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
}
