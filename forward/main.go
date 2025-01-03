package main

import (
	"context"
	_ "embed"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/injoyai/proxy/core"
	"github.com/injoyai/proxy/forward"
)

//go:embed index.html
var index string

var (
	DB    *xorms.Engine
	Cache = maps.NewSafe()
)

func init() {
	var err error
	DB, err = sqlite.NewXorm(oss.UserInjoyDir("/forward/database/forward.db"))
	logs.PanicErr(err)

	err = DB.Sync(&Forward{})
	logs.PanicErr(err)

	all := []*Forward(nil)
	err = DB.Find(&all)
	logs.PanicErr(err)

	for _, f := range all {
		f.init()
		Cache.Set(conv.String(f.ID), f)
	}
}

func main() {
	tray.Run(
		tray.WithIco(ico),
		tray.WithHint("端口转发"),
		func(s *tray.Tray) {
			s.AddMenu().SetName("配置").SetIcon(tray.IconSetting).OnClick(func(m *tray.Menu) {
				logs.PrintErr(gui())
			})
		},
		tray.WithStartup(),
		tray.WithSeparator(),
		tray.WithExit(),
	)
}

func gui() error {

	return lorca.Run(&lorca.Config{
		Width:  800,
		Height: 600,
		Index:  index,
	}, func(app lorca.APP) error {

		app.Bind("fnGetForwardAll", func() []*Forward {
			ls := []*Forward(nil)
			err := DB.Desc("ID").Find(&ls)
			if err != nil {
				return nil
			}
			for _, v := range ls {
				v.resp()
			}
			return ls
		})

		app.Bind("fnAddForward", func(name, port, address string) error {
			f := &Forward{
				Name:    name,
				Port:    conv.Int(port),
				Address: address,
			}
			_, err := DB.Insert(f)
			if err == nil {
				f.init()
				Cache.Set(conv.String(f.ID), f)
			}
			return err
		})

		app.Bind("fnUpdateForward", func(id, name, port, address string) error {
			f := &Forward{
				ID:      conv.Int64(id),
				Name:    name,
				Port:    conv.Int(port),
				Address: address,
			}
			_, err := DB.Where("ID=?", id).Cols("Name", "Port", "Address").Update(f)
			if err == nil {
				val, ok := Cache.Get(id)
				if ok {
					val.(*Forward).r.Disable()
					f.Enable = val.(*Forward).Enable
				}
				f.init()
				Cache.Set(id, f)
			}
			return err
		})

		app.Bind("fnDelForward", func(id string) error {
			_, err := DB.ID(id).Delete(&Forward{})
			if err == nil {
				val, ok := Cache.Get(id)
				if ok {
					val.(*Forward).r.Disable()
				}
				Cache.Del(id)
			}
			return err
		})

		app.Bind("fnStartForward", func(id string) error {
			_, err := DB.ID(id).Cols("Enable").Update(&Forward{
				Enable: true,
			})
			if err != nil {
				logs.Err(err)
				return err
			}
			f, ok := Cache.Get(id)
			if ok {
				f.(*Forward).enable()
			}
			return nil
		})

		app.Bind("fnStopForward", func(id string) error {
			_, err := DB.ID(id).Cols("Enable").Update(&Forward{
				Enable: false,
			})
			if err != nil {
				logs.Err(err)
				return err
			}
			f, ok := Cache.Get(id)
			if ok {
				f.(*Forward).disable()
			}
			return nil
		})

		return nil
	})

}

type Forward struct {
	ID      int64  `json:"id"`               //主键
	Name    string `json:"name"`             //名称
	Port    int    `json:"port"`             //监听端口
	Address string `json:"address"`          //转发地址
	Enable  bool   `json:"enable"`           //是否启用
	Running bool   `json:"running" xorm:"-"` //是否运行
	Error   string `json:"error" xorm:"-"`   //错误信息
	r       *chans.Rerun
}

func (this *Forward) resp() {
	val, ok := Cache.Get(conv.String(this.ID))
	if ok {
		this.Running = val.(*Forward).Running
		this.Error = val.(*Forward).Error
	}
}

func (this *Forward) init() {
	this.r = chans.NewRerun(func(ctx context.Context) {
		f := &forward.Forward{
			Listen:  core.NewListenTCP(this.Port),
			Forward: core.NewDialTCP(this.Address),
		}
		this.Running = true
		err := f.Run(ctx)
		this.Running = false
		this.Error = conv.String(err)
	})
	if this.Enable {
		this.r.Enable()
	}
}

func (this *Forward) enable() {
	this.Enable = true
	this.r.Enable()
}

func (this *Forward) disable() {
	this.Enable = false
	this.r.Disable()
}
