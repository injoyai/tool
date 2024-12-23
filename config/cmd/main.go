package main

import (
	"github.com/injoyai/tool/config"
)

func main() {

	config.GUI(config.NewConfig("config.json", nil)) // testdata.DefaultNatures))
}
