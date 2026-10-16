package main

import (
	"github.com/injoyai/logs"
	"github.com/injoyai/tool/crawl-images/baidu"
)

func main() {
	err := baidu.Crawl("玩水俯拍", "images", 30*100)
	logs.PrintErr(err)
}
