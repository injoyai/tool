package main

import (
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/frame/fiber"
	"github.com/injoyai/logs"
)

func init() {
	logs.SetFormatter(logs.TimeFormatter)
	cfg.Init(
		cfg.WithEnv(),
		cfg.WithFlag(
			&cfg.Flag{Name: "port", Usage: "监听端口"},
			&cfg.Flag{Name: "root", Usage: "资源目录"},
		),
		cfg.WithFile("./config/config.yaml"),
	)
}

func main() {
	port := cfg.GetInt("port", 80)
	root := cfg.GetString("root", "./")

	s := fiber.Default()
	s.SetPort(port)
	s.Use(fiber.WithStatic(root))
	err := s.Run()
	logs.Err(err)
}
