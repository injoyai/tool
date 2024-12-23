package main

import "github.com/injoyai/goutil/oss/tray"

func Run(f func() error) {
	tray.Run(
		func(s *tray.Tray) {
			go func() {
				s.SetHint("状态: 运行中\n端口: 9000\n地址: 192.168.192.2:9000")
				err := f()
				s.SetHint("状态: " + err.Error() + "\n端口: 9000\n地址: 192.168.192.2:9000")
			}()
		},
		tray.WithIco(nil),
		tray.WithLabel("v1.0.1"),
		tray.WithStartup(),
		tray.WithExit(),
	)
}
