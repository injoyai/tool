package timer

import "github.com/injoyai/goutil/notice"

func Notice(msg string) error {
	return notice.DefaultWindows.Publish(&notice.Message{
		Content: msg,
	})
}
