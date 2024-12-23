package config

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/lorca"
)

//go:embed index.html
var html string

func GUI(cfg *Config) error {

	return lorca.Run(&lorca.Config{
		Width:  720,
		Height: 860,
		Index:  html,
	}, func(app lorca.APP) error {

		configs := cfg.Get()
		//加载配置数据
		app.Eval(fmt.Sprintf(`loadConfig(%s)`, conv.String(configs)))

		app.Bind("loading", func() {
			configs := cfg.Get()
			//加载配置数据
			app.Eval(fmt.Sprintf(`loadConfig(%s)`, conv.String(configs)))
		})

		//获取保存数据
		app.Bind("saveToFile", func(config interface{}) {
			if err := cfg.Save(conv.GMap(config)); err != nil {
				app.Eval(fmt.Sprintf(`notice("%v");`, err))
			} else {
				app.Eval(`notice("保存成功");`)
			}
		})

		return nil
	})
}
