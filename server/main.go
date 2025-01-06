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
		{"version": "v2.1", "desc": "默认菜单可关闭"},
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

	tcp := chans.NewRerun(ss.RunTCP)
	http := chans.NewRerun(ss.RunHTTP)
	tcp.Enable(cfg.GetBool("tcp.enable"))
	http.Enable(cfg.GetBool("http.enable"))

	tray.Run(
		func(s *tray.Tray) {
			s.SetHint("In Server")
			s.SetIco(IconI)
			m := s.AddMenu().SetName("版本: " + Version).SetIcon(IconVersion)
			for _, v := range VersionHistory {
				m.AddMenu().SetName(v["version"].(string) + "  " + v["desc"].(string)).Disable()
			}

			s.AddMenu().SetName("服务配置").SetIcon(IconSetting).OnClick(func(m *tray.Menu) {
				config.GUI(config.New(Filename, config.Natures{
					{Name: "TCP", Key: "tcp", Type: "object2", Value: config.Natures{
						{Name: "启用", Key: "enable", Type: "bool"},
						{Name: "端口", Key: "port"},
					}},
					{Name: "HTTP", Key: "http", Type: "object2", Value: config.Natures{
						{Name: "启用", Key: "enable", Type: "bool"},
						{Name: "端口", Key: "port"},
					}},
					{Name: "默认菜单", Key: "default_menu", Type: "bool"},
					{Name: "自定义菜单", Key: "menu", Type: "object"},
				}).SetWidthHeight(800, 600).OnSaved(func(m *conv.Map) {
					tcp.Enable(m.GetBool("tcp.enable"))
					http.Enable(m.GetBool("http.enable"))
				}))
			})
			if cfg.GetBool("default_menu") {
				s.AddMenu().SetName("全局配置").SetIcon(IconSetting).OnClick(func(m *tray.Menu) {
					shell.Start("in global gui")
				})
				s.AddMenu().SetName("定时任务").SetIcon(IconTimer).OnClick(func(m *tray.Menu) {
					shell.Start("in open timer")
				})
				s.AddMenu().SetName("消息通知").SetIcon(IconNotice).OnClick(func(m *tray.Menu) {
					shell.Start("in open notice_client")
				})
				s.AddMenu().SetName("文件服务").OnClick(func(m *tray.Menu) {
					shell.Start("in open hfs")
				})
			}

			//加载自定义菜单
			s.AddSeparator()
			for k, v := range cfg.GetMap("menu") {
				cmd := conv.String(v)
				s.AddMenu().SetName(k).OnClick(func(m *tray.Menu) {
					shell.Run(cmd)
				})
			}

			s.AddSeparator()
			tray.WithStartup(tray.Name("开机自启"))(s)
			mHTTP := s.AddMenuCheck().SetChecked(cfg.GetBool("http.enable"))
			mHTTP.SetName("HTTP: " + conv.String(cfg.GetInt("http.port"))).OnClick(func(m *tray.Menu) {
				x := cfg.WithFile(Filename).(*conv.Map).Set("http.enable", !m.Checked())
				oss.New(Filename, x.String())
				http.Enable(!m.Checked())
				mHTTP.SetChecked(!m.Checked())
			})
			mTCP := s.AddMenuCheck().SetChecked(cfg.GetBool("tcp.enable"))
			mTCP.SetName("TCP  : " + conv.String(cfg.GetInt("tcp.port"))).OnClick(func(m *tray.Menu) {
				x := cfg.WithFile(Filename).(*conv.Map).Set("tcp.enable", !m.Checked())
				oss.New(Filename, x.String())
				tcp.Enable(!m.Checked())
				mTCP.SetChecked(!m.Checked())
			})

			s.AddSeparator()
			s.AddMenu().SetName("退出").SetIcon(IconExit).OnClick(func(m *tray.Menu) {
				s.Close()
			})
		},
	)

}
