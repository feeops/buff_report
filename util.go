package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	replace = []string{
		"Key: ",
		":",
		"：",
		" ",
		"'",
		"$",
		"(",
		"¥",
		")",
		"USD",
	}
)

func waitExit() {
	fmt.Println("请按任意键退出(此操作会同时关闭程序)")
	fmt.Scanln()
	os.Exit(0)
}

func cleanStr(oldStr string) string {
	oldStr = strings.TrimSpace(oldStr)

	for _, item := range replace {
		oldStr = strings.ReplaceAll(oldStr, item, "")
	}

	oldStr = strings.TrimSpace(oldStr)

	return oldStr
}
