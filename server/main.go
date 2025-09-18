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
		{"version": "v2.4", "desc": "增加通知服务,修改tcp为udp"},
		{"version": "v2.3", "desc": "配合in工具修改名字为i"},
		{"version": "v2.2", "desc": "增加文件服务,重新启动,版本升级菜单"},
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

	udp := chans.NewRerun(ss.RunUDP)
	http := chans.NewRerun(ss.RunHTTP)
	udp.Enable(cfg.GetBool("tcp.enable"))
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
					{Name: "UDP", Key: "udp", Type: "object2", Value: config.Natures{
						{Name: "启用", Key: "enable", Type: "bool"},
						{Name: "端口", Key: "port"},
					}},
					{Name: "HTTP", Key: "http", Type: "object2", Value: config.Natures{
						{Name: "启用", Key: "enable", Type: "bool"},
						{Name: "端口", Key: "port"},
					}},
					{Name: "默认菜单", Key: "menu_default"},
					{Name: "自定义菜单", Key: "menu", Type: "object"},
				}).SetWidthHeight(800, 800).OnSaved(func(m *conv.Map) {
					logs.Debug(m.String())
					udp.Enable(m.GetBool("udp.enable"))
					http.Enable(m.GetBool("http.enable"))
					shell.Start("i open server")
				}))
			})
			if cfg.GetBool("menu_default", true) {
				s.AddMenu().SetName("全局配置").SetIcon(IconSetting).OnClick(func(m *tray.Menu) {
					shell.Start("i global gui")
				})
				s.AddMenu().SetName("定时任务").SetIcon(IconTimer).OnClick(func(m *tray.Menu) {
					shell.Start("i open timer")
				})
				s.AddMenu().SetName("消息通知").SetIcon(IconNotice).OnClick(func(m *tray.Menu) {
					shell.Start("i open notice_client")
				})
				s.AddMenu().SetName("文件服务").OnClick(func(m *tray.Menu) {
					shell.Start("i open hfs")
				})
				s.AddMenu().SetName("重新启动").OnClick(func(m *tray.Menu) {
					shell.Start("i open server")
				})
				s.AddMenu().SetName("版本升级").OnClick(func(m *tray.Menu) {
					shell.Start("i open server upgrade")
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
			mTCP := s.AddMenuCheck().SetChecked(cfg.GetBool("udp.enable"))
			mTCP.SetName("UDP  : " + conv.String(cfg.GetInt("udp.port"))).OnClick(func(m *tray.Menu) {
				x := cfg.WithFile(Filename).(*conv.Map).Set("udp.enable", !m.Checked())
				oss.New(Filename, x.String())
				udp.Enable(!m.Checked())
				mTCP.SetChecked(!m.Checked())
			})

			s.AddSeparator()
			s.AddMenu().SetName("退出").SetIcon(IconExit).OnClick(func(m *tray.Menu) {
				s.Close()
			})
		},
	)

}
