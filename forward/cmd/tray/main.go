package main

import (
	"fmt"

	"github.com/injoyai/forward"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
)

func main() {
	port := 40073
	db := oss.UserInjoyDir("/forward/database/forward.db")
	go func() { logs.PanicErr(forward.Run(port, db)) }()
	tray.Run(
		tray.WithIco(ico),
		tray.WithHint("端口转发"),
		func(s *tray.Tray) {
			s.AddMenu().SetName("配置").SetIcon(tray.IconSetting).OnClick(func(m *tray.Menu) {
				lorca.Run(&lorca.Config{
					Width:  920,
					Height: 600,
					Index:  fmt.Sprintf("http://localhost:%d", port),
				})
			})
		},
		tray.WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
	)
}
