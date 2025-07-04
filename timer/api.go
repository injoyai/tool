package timer

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/frame/fiber"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/script"
	"github.com/injoyai/goutil/script/js"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"net"
	"xorm.io/xorm"
)

//go:embed index.html
var index []byte

const Startup = "@startup"

var (
	DB     *xorms.Engine
	Script = js.NewPool(10, script.WithObject, script.WithFunc)
	Corn   = task.New().Start()
	WS     = maps.NewGeneric[*fiber.Websocket, *fiber.Websocket]()
)

func _init(filename string) (err error) {
	logs.SetWriter(logs.Stdout)

	DB, err = sqlite.NewXorm(filename)
	if err != nil {
		return err
	}
	if err = DB.Sync2(new(Timer)); err != nil {
		return err
	}

	Script.SetFunc("start", func(args *script.Args) (interface{}, error) {
		err := shell.Start2(args.GetString(1))
		return nil, err
	})

	Script.SetFunc("ping", func(args *script.Args) (interface{}, error) {
		result, err := ip.Ping(args.GetString(1), args.Get(2).Second(1))
		logs.Debug(result, err)
		return result.String(), err
	})

	Script.SetFunc("notice", func(args *script.Args) (interface{}, error) {
		msg := args.GetString(1)
		target := args.GetString(2)

		switch target {
		case "popup", "pop", "":
			return nil, notice.DefaultWindows.Publish(&notice.Message{
				Content: msg,
				Target:  target,
			})

		case "client":
			WS.Range(func(key *fiber.Websocket, value *fiber.Websocket) bool {
				value.Write([]byte(msg))
				return true
			})
			return nil, nil

		default:

			x := http.Url(target).
				SetContentType("application/json").
				Debug().
				SetBody(msg)
			err := x.Post().Err()
			return nil, err

		}

	})

	Script.SetFunc("dial", func(args *script.Args) (interface{}, error) {
		network := args.GetString(1)
		address := args.GetString(2)
		timeout := args.Get(3).Second(2)
		c, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			return nil, err
		}
		c.Close()
		return "成功", nil
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

	Script.SetFunc("print", func(args *script.Args) (interface{}, error) {
		fmt.Println(args.Interfaces()...)
		return nil, nil
	})

	data := []*Timer(nil)
	DB.Find(&data)
	for i := range data {
		v := data[i]
		logs.Debug(v)
		if !v.Enable {
			continue
		}
		exec := func(v *Timer) {
			logs.Trace(v.ExecText())
			if _, err := Script.Exec(v.Content); err != nil {
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   fmt.Sprintf("定时任务[%s]执行错误:", v.Name),
					Content: err.Error(),
				})
			}
		}
		if v.Cron == Startup {
			exec(v)
			continue
		}
		Corn.SetTask(conv.String(v.ID), v.Cron, func() { exec(v) })
	}

	return nil
}

func Run(port int, filename string) error {
	if err := _init(filename); err != nil {
		return err
	}
	s := fiber.Default()
	s.SetPort(port)
	s.GET("/", func(c fiber.Ctx) { c.Html200(index) })
	s.Group("/api", func(g fiber.Grouper) {
		g.GET("/timer/all", GetTimerAll)    //列表
		g.POST("/timer", PostTimer)         //新建
		g.PUT("/timer", PutTimer)           //修改
		g.DELETE("/timer", DelTimer)        //删除
		g.PUT("/timer/enable", EnableTimer) //启用/禁用
		g.ALL("/notice/ws", NoticeWS)       //通知-websocket
	})
	return s.Run()
}

func GetTimerAll(c fiber.Ctx) {
	data := []*Timer(nil)
	err := DB.Find(&data)
	c.CheckErr(err)
	for _, v := range data {
		v.Resp(Corn.GetTask(conv.String(v.ID)))
	}
	c.Succ(data)
}

func PostTimer(c fiber.Ctx) {
	t := &Timer{}
	c.Parse(t)
	_, err := DB.Insert(t)
	c.CheckErr(err)
	if t.Enable && t.Cron != Startup {
		err = Corn.SetTask(conv.String(t.ID), t.Cron, func() {
			if _, err = Script.Exec(t.Content); err != nil {
				t.ExecErr = err.Error()
			}
		})
		c.CheckErr(err)
	}
	c.Succ(nil)
}

func PutTimer(c fiber.Ctx) {
	t := &Timer{}
	c.Parse(t)

	_, err := DB.ID(t.ID).AllCols().Update(t)
	c.CheckErr(err)

	Corn.DelTask(conv.String(t.ID))
	if t.Enable && t.Cron != Startup {
		err = Corn.SetTask(conv.String(t.ID), t.Cron, func() {
			if _, err = Script.Exec(t.Content); err != nil {
				t.ExecErr = err.Error()
			}
		})
		c.CheckErr(err)
	}
	c.Succ(nil)
}

func DelTimer(c fiber.Ctx) {
	id := c.Get("id")
	_, err := DB.ID(id).Delete(new(Timer))
	c.CheckErr(err)
	Corn.DelTask(id)
	c.Succ(nil)
}

func EnableTimer(c fiber.Ctx) {
	t := new(Timer)
	c.Parse(t)

	DB.SessionFunc(func(session *xorm.Session) error {
		if _, err := session.ID(t.ID).Cols("Enable").Update(t); err != nil {
			return err
		}
		if t.Enable && t.Cron != Startup {
			if err := Corn.SetTask(conv.String(t.ID), t.Cron, func() {
				if _, err := Script.Exec(t.Content); err != nil {
					t.ExecErr = err.Error()
				}
			}); err != nil {
				return err
			}
		} else {
			Corn.DelTask(conv.String(t.ID))
		}
		return nil
	})

}

func NoticeWS(c fiber.Ctx) {
	c.Websocket(func(conn *fiber.Websocket) {
		WS.Set(conn, conn)
		defer WS.Del(conn)
		<-conn.Done()
	})
}
