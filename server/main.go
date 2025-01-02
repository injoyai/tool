package main

import (
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/logs"
	"github.com/injoyai/tool/config"
	"github.com/injoyai/tool/server/service"
)

const (
	Fail = 500
	Succ = 200
)

var (
	Filename       = oss.UserInjoyDir("/server/config.yaml")
	Version        = VersionHistory[0]["version"].(string)
	VersionHistory = []g.Map{
		{"version": "v2.0", "desc": "独立一个项目,方便修改"},
	}
)

func main() {

	logs.SetFormatterWithTime()
	cfg.Init(cfg.WithFile(Filename))

	ss := &service.Server{
		Filename: Filename,
		Version:  Version,
		Fail:     Fail,
		Succ:     Succ,
	}

	rerun := chans.NewRerun(ss.Run)
	rerun.Enable()

	tray.Run(
		func(s *tray.Tray) {
			s.SetHint("In Server")
			s.SetIco(IconI)
			m := s.AddMenu().SetName("版本: " + Version).SetIco(IconVersion)
			for _, v := range VersionHistory {
				m.AddMenu().SetName(v["version"].(string) + "  " + v["desc"].(string)).Disable()
			}

			s.AddMenu().SetName("服务配置").SetIco(IconSetting).OnClick(func(m *tray.Menu) {
				config.GUI(config.New(Filename, config.Natures{
					{Name: "TCP端口", Key: "tcp_port"},
					{Name: "HTTP端口", Key: "http_port"},
					{Name: "自定义菜单", Key: "custom_menu", Type: "object"},
				}).OnSaved(func(m *conv.Map) {
					rerun.Rerun()
				}))
			})
			s.AddMenu().SetName("全局配置").SetIco(IconSetting).OnClick(func(m *tray.Menu) {
				shell.Start("in global gui")
			})
			s.AddMenu().SetName("定时任务").SetIco(IconTimer).OnClick(func(m *tray.Menu) {
				shell.Start("in open timer")
			})
			s.AddMenu().SetName("消息通知").SetIco(IconNotice).OnClick(func(m *tray.Menu) {
				shell.Start("in open notice_client")
			})

			//加载自定义菜单
			for k, v := range cfg.GetMap("custom_menu") {
				cmd := conv.String(v)
				s.AddMenu().SetName(k).OnClick(func(m *tray.Menu) {
					shell.Run(cmd)
				})
			}

			s.AddSeparator()
			tray.WithStartup()(s)
			s.AddSeparator()
			s.AddMenu().SetName("退出").SetIco(IconExit).OnClick(func(m *tray.Menu) {
				s.Close()
			})
		},
	)

}
