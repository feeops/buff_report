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
	}
	_ = f.SetSheetRow("Sheet1", "A1", &titleSlice)

	fmt.Printf("buff指定时间内历史总记录数是%d条,无法匹配上的会自动跳过\n", len(buffItems))

	for k, item := range buffItems {

		buffPrice := item.buffPrice
		assetID := item.assetID
		secondAssetID := secondMap[assetID]
		steamPrice := thirdMap[secondAssetID]
		marketHashName := classInstanceName[assetName[item.assetID]]

		if len(marketHashName) == 0 {
			break
		}

		ratio, _ := decimal.
			NewFromFloat((buffPrice*(1-0.025)*(1-0.01) - steamPrice*exchangeRate) * 100 / (steamPrice * exchangeRate)).
			Round(2).Float64()
		profit, _ := decimal.NewFromFloat(buffPrice*(1-0.025)*(1-0.01) - steamPrice*exchangeRate).Round(2).Float64()
		data := []interface{}{
			item.marketName,
			marketHashName,
			steamPrice,
			exchangeRate,
			exchangeRate * steamPrice,
			buffPrice,
			profit,
			ratio,
			item.tradeTime,
			item.buffLink,
		}

		f.SetSheetRow("sheet1", fmt.Sprint("A", k+2), &data)
	}
	// 写入数据
	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
}
