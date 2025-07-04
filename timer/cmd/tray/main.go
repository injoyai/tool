package main

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
)

func main() {
	host := cfg.GetString("host", "127.0.0.1:8080")
	go listen(fmt.Sprintf("ws://%s/api/notice/ws", host))
	tray.Run(
		tray.WithIco(IcoTimer),
		tray.WithHint("定时任务"),
		func(s *tray.Tray) {
			x := s.AddMenu().SetName("配置").SetIcon(IcoMenuTimer)
			x.OnClick(func(m *tray.Menu) {
				lorca.Run(&lorca.Config{
					Width:  930,
					Height: 680,
					Index:  host,
				})
			})
		},
		tray.WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
	)
}

func listen(url string) {
	dial.RedialWebsocket(url, func(c *client.Client) {
		c.OnDealMessage = func(c *client.Client, msg ios.Acker) {
			m := &notice.Message{}
			if err := json.Unmarshal(msg.Payload(), m); err != nil {
				logs.Err(err)
				return
			}
			err := notice.DefaultWindows.Publish(m)
			logs.PrintErr(err)
		}
	})
}
