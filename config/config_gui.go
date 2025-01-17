package config

import (
	_ "embed"
	"github.com/injoyai/lorca"
)

type APP = lorca.APP

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
		//绑定函数
		app.Bind("goGetConfig", cfg.Get)
		app.Bind("goSaveConfig", cfg.Save)
		//加载配置数据
		app.Eval(`loadConfig()`)
		return nil
	})
}
