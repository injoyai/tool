package config

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/frame/mux"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"io/fs"
	"net"
	"net/http"
)

//go:embed index.html
var html string

//go:embed dist/*
var dist embed.FS

func Run(filename string) error {

	s := mux.New()
	s.SetPort(0)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	web, err := fs.Sub(dist, "dist")
	if err != nil {
		return err
	}
	go http.Serve(l, http.FileServer(http.FS(web)))

	return lorca.Run(&lorca.Config{
		Width:  720,
		Height: 860,
		Index:  "http://" + l.Addr().String() + "/index.html",
	}, func(app lorca.APP) error {

		configs := GetConfigs()

		//加载配置数据
		app.Eval(fmt.Sprintf(`loadConfig(%s)`, conv.String(configs)))

		//获取保存数据
		app.Bind("saveToFile", func(config interface{}) {
			fmt.Println(config)
			if err := SaveConfigs(conv.GMap(config)); err != nil {
				logs.Err(err)
				app.Eval(fmt.Sprintf(`notice("%v");`, err))
			} else {
				app.Eval(`notice("保存成功");`)
			}
		})

		return nil
	})
}
