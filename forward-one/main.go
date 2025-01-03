package main

import (
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/proxy/core"
	"github.com/injoyai/proxy/forward"
)

const Version = "v1.0.1"

var configPath = oss.UserInjoyDir("/forward/config/config.yaml")

func main() {
	cfg.Init(
		cfg.WithFile(configPath),
		cfg.WithFlag(
			&cfg.Flag{Name: "port", Default: 9000, Usage: "本地监听端口"},
			&cfg.Flag{Name: "address", Default: "192.168.192.2:9000", Usage: "转发地址"},
		),
	)
	f := &forward.Forward{
		Listen:  core.NewListenTCP(cfg.GetInt("port")),
		Forward: core.NewDialTCP(cfg.GetString("address")),
	}

	Run(f)
}
