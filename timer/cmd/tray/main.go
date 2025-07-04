package main

import (
	"fmt"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/lorca"
	"github.com/injoyai/tool/timer"
)

func main() {
	port := 60074
	db := oss.UserInjoyDir("/timer/database/timer.db")
	go timer.Run(port, db)
	tray.Run(
		tray.WithIco(IcoTimer),
		tray.WithHint("定时任务"),
		func(s *tray.Tray) {
			x := s.AddMenu().SetName("配置").SetIcon(IcoMenuTimer)
			x.OnClick(func(m *tray.Menu) {
				lorca.Run(&lorca.Config{
					Width:  930,
					Height: 680,
					Index:  fmt.Sprintf("http://localhost:%d", port),
				})
			})
		},
		tray.WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
	)
}
