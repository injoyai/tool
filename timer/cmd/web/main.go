package main

import (
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/tool/timer"
)

func main() {
	timer.Run(
		cfg.GetInt("port", 8078),
		cfg.GetString("db", "./data/database/timer.db"),
	)
}
