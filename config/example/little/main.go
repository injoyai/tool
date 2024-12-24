package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/tool/config"
	"log"
)

func main() {
	config.GUI(config.New("./example/little/config.yaml", config.Natures{
		{Key: "port", Name: "监听端口"},
		{Key: "forward", Name: "转发地址"},
	}).SetWidthHeight(720, 350).OnSaved(func(m *conv.Map) {
		log.Println("保存成功")
		fmt.Println(m.String())
	}))
}
