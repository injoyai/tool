package main

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/goutil/script"
	"github.com/injoyai/goutil/script/js"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"net"
	"xorm.io/xorm"
)

//go:embed index.html
var html string

var (
	DB, _  = sqlite.NewXorm(oss.UserInjoyDir("/timer/timer.db"))
	Script = js.NewPool(10, script.WithBaseFunc)
	Corn   = task.New().Start()
)

func init() {
	logs.SetWriter(logs.Stdout)
	DB.Sync2(new(Timer))
	data := []*Timer(nil)
	DB.Find(&data)
	for i := range data {
		v := data[i]
		logs.Debug(v)
		if !v.Enable {
			continue
		}
		Corn.SetTask(conv.String(v.ID), v.Cron, func() {
			logs.Trace(v.ExecText())
			if _, err := Script.Exec(v.Content); err != nil {
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   fmt.Sprintf("定时任务[%s]执行错误:", v.Name),
					Content: err.Error(),
				})
			}
		})
	}
	Script.SetFunc("ping", func(args *script.Args) (interface{}, error) {
		result, err := ip.Ping(args.GetString(1), args.Get(2).Second(1))
		return result.String(), err
	})

	Script.SetFunc("notice", func(args *script.Args) (interface{}, error) {
		target := args.GetString(2)
		notice.DefaultWindows.Publish(&notice.Message{
			Content: args.GetString(1),
			Target:  target,
		})
		return nil, nil
	})
	Script.SetFunc("dialTCP", func(args *script.Args) (interface{}, error) {
		address := args.GetString(1)
		timeout := args.Get(2).Second(2)
		c, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return nil, err
		}
		c.Close()
		return "成功", nil
	})
	Script.SetFunc("notice_wechat_friend", func(args *script.Args) (interface{}, error) {
		address := args.GetString(1)
		target := args.GetString(2)
		msg := args.GetString(3)
		err := http.Url(fmt.Sprintf("http://%s/api/notice", address)).
			SetToken("147258369").
			SetContentType("application/json").
			SetBody(g.Map{
				"output":  []string{"wechat:friend:" + target},
				"content": msg,
			}).Debug().Post().Err()
		return nil, err
	})
	Script.SetFunc("notice_wechat_group", func(args *script.Args) (interface{}, error) {
		address := args.GetString(1)
		target := args.GetString(2)
		msg := args.GetString(3)
		err := http.Url(fmt.Sprintf("http://%s/api/notice", address)).
			SetToken("147258369").
			SetContentType("application/json").
			SetBody(g.Map{
				"output":  []string{"wechat:group:" + target},
				"content": msg,
			}).Debug().Post().Err()
		return nil, err
	})

}

func main() {

	tray.Run(
		tray.WithIco(IcoTimer),
		func(s *tray.Tray) {

			x := s.AddMenu().SetName("显示").SetIco(IcoMenuTimer)
			defer func() {
				x.ClickedCh <- struct{}{}
			}()
			x.OnClick(func(m *tray.Menu) {

				err := lorca.Run(&lorca.Config{
					Width:  1080,
					Height: 1080,
					Index:  html,
				}, func(app lorca.APP) error {

					Script.SetFunc("print", func(args *script.Args) (interface{}, error) {
						s := fmt.Sprint(args.Interfaces()...)
						app.Eval(fmt.Sprintf(`notice("%s")`, s))
						return nil, nil
					})

					t := &handler{app: app}

					app.Bind("addTimer", func(name, cron, content string, enable bool) {
						logs.Debug("addTimer")
						defer t.Refresh()
						if err := t.AddTimer(name, cron, content, enable); err != nil {
							app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
							return
						}
					})

					app.Bind("updateTimer", func(id, name, cron, content string) {
						logs.Debug("updateTimer")
						defer t.Refresh()
						if err := t.UpdateTimer(id, name, cron, content); err != nil {
							app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
							return
						}
					})

					app.Bind("enableTimer", func(id string, enable bool) {
						logs.Debug("enableTimer")
						defer t.Refresh()
						if err := t.EnableTimer(id, enable); err != nil {
							app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
							return
						}
					})

					app.Bind("delTimer", func(id string) {
						logs.Debug("delTimer")
						defer t.Refresh()
						if err := t.DelTimer(id); err != nil {
							app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
							return
						}
					})

					app.Bind("refresh", func() { t.Refresh() })

					t.Refresh()
					return nil
				})
				logs.Debug(err)
			})

			tray.WithStartup()(s)

			s.AddMenu().SetName("退出").SetIco(IcoMenuQuit).OnClick(func(m *tray.Menu) {
				s.Close()
			})

		},
	)

}

type handler struct {
	app lorca.APP
}

func (this *handler) Refresh() {
	logs.Debug("Refresh")
	this.app.Eval("notice('刷新')")
	data := []*Timer(nil)
	if err := DB.Find(&data); err != nil {
		this.app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
		return
	}
	this.app.Eval("clearTimer()")
	for _, v := range data {
		next := ""
		if t := Corn.GetTask(conv.String(v.ID)); t != nil {
			next = t.Next.Format("2006-01-02 15:04:05")
		}
		this.app.Eval(fmt.Sprintf(`loadingTimer(%d,'%s','%s','%s',%t,'%s')`, v.ID, v.Name, v.Cron, v.Content, v.Enable, next))
	}
}

func (this *handler) AddTimer(name, cron, content string, enable bool) error {
	t := &Timer{
		Name:    name,
		Cron:    cron,
		Content: content,
		Enable:  enable,
	}
	if _, err := DB.Insert(t); err != nil {
		return err
	}
	if t.Enable {
		if err := Corn.SetTask(conv.String(t.ID), t.Cron, func() {
			if _, err := Script.Exec(t.Content); err != nil {
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   fmt.Sprintf("定时任务[%s]执行错误:", t.Name),
					Content: err.Error(),
				})
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

func (this *handler) UpdateTimer(id, name, cron, content string) error {
	t := new(Timer)
	if _, err := DB.ID(id).Get(t); err != nil {
		return err
	}
	t.Name = name
	t.Cron = cron
	t.Content = content

	if _, err := DB.ID(id).AllCols().Update(t); err != nil {
		return err
	}

	Corn.DelTask(id)
	if t.Enable {
		if err := Corn.SetTask(id, t.Cron, func() {
			if _, err := Script.Exec(t.Content); err != nil {
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   fmt.Sprintf("定时任务[%s]执行错误:", t.Name),
					Content: err.Error(),
				})
			}
		}); err != nil {
			return err
		}
	}

	return nil
}

func (this *handler) EnableTimer(id string, enable bool) error {
	t := new(Timer)
	if _, err := DB.ID(id).Get(t); err != nil {
		return err
	}
	t.Enable = enable
	logs.Debugf("[%s][%s][%d:%s] %s\n", conv.SelectString(t.Enable, "启用", "禁用"), t.Cron, t.ID, t.Name, t.Content)
	return DB.SessionFunc(func(session *xorm.Session) error {
		if _, err := session.ID(id).AllCols().Update(t); err != nil {
			return err
		}
		if enable {
			if err := Corn.SetTask(id, t.Cron, func() {
				if _, err := Script.Exec(t.Content); err != nil {
					notice.DefaultWindows.Publish(&notice.Message{
						Title:   fmt.Sprintf("定时任务[%s]执行错误:", t.Name),
						Content: err.Error(),
					})
				}
			}); err != nil {
				return err
			}
		} else {
			Corn.DelTask(id)
		}
		return nil
	})
}

func (this *handler) DelTimer(id string) error {
	_, err := DB.ID(id).Delete(new(Timer))
	if err != nil {
		return err
	}
	Corn.DelTask(id)
	return nil
}

type Timer struct {
	ID      int64
	Name    string
	Cron    string
	Content string
	Enable  bool
}

func (this *Timer) String() string {
	return fmt.Sprintf("[%s][%s][%02d:%s] %s", conv.SelectString(this.Enable, "启用", "禁用"), this.Cron, this.ID, this.Name, this.Content)
}

func (this *Timer) ExecText() string {
	return fmt.Sprintf("[执行][%s][%02d:%s] %s", this.Cron, this.ID, this.Name, this.Content)
}
