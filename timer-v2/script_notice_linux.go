package timer

import (
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/notice/pkg/push"
	"github.com/injoyai/notice/pkg/push/serverchan"
)

func Notice(title, msg string) error {
	return serverchan.New(cfg.GetString("notice.serverchan.key")).Push(&push.Message{
		Title:   title,
		Content: msg,
	})
}
