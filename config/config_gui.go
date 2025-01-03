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
	if cfg.Width <= 0 {
		cfg.Width = 720
	}
	if cfg.Height <= 0 {
		cfg.Height = 480
	}

	return lorca.Run(&lorca.Config{
		Width:  cfg.Width,
		Height: cfg.Height,
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
			err := cfg.Save(conv.GMap(config))
			if err != nil {
				app.Eval(fmt.Sprintf(`notice("%v");`, err))
			} else {
				app.Eval(`notice("保存成功");`)
				if cfg.OnSaved != nil {
					cfg.OnSaved(cfg.m)
				}
			}

		})

		return nil
	})
}
