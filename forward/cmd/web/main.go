package main

import (
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/forward"
	"github.com/injoyai/frame/fiber"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
)

func main() {
	logs.Err(forward.Run(
		cfg.GetInt("port", 8079),
		cfg.GetString("database", "./data/forward.db"),
		auth,
	))
}

var token string

func auth(c fiber.Ctx) {
	if cfg.GetBool("auth") {
		switch c.Path() {
		case "/api/login":
			if c.GetString("password") != cfg.GetString("password") {
				c.Fail("密码错误")
			}
			token = g.RandString(8)
			c.Succ(token)
		default:
			if c.Get("Authorization") != token {
				c.Unauthorized()
			}
		}
	}
	c.Next()
}
