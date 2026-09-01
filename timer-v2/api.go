package timer

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"sync"
	"time"

	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/frame/fbr"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs/v2"
	"xorm.io/xorm"
)

//go:embed web
var webFS embed.FS

const Startup = "@startup"

// wsClient 包装 websocket 连接与其写锁。
// fasthttp/websocket 禁止并发写同一连接(会 panic)，需用互斥锁序列化 WriteJSON 等写操作。
type wsClient struct {
	conn *fbr.Websocket
	mu   sync.Mutex
}

var (
	DB     *xorms.Engine
	Script = newScriptEngine()
	Corn   = task.New().Start()
	WS     = maps.NewGeneric[*fbr.Websocket, *wsClient]()
)

func _init(filename string) (err error) {
	logs.SetWriter(logs.Stdout)

	DB, err = sqlite.NewXorm(filename)
	if err != nil {
		return err
	}
	// xorm 默认 DatabaseTZ=time.UTC，导致 created 字段存 UTC 时间(比本地少 8 小时)，
	// 改为本地时区，使 Log.CreatedAt 等时间字段存本地时间。
	DB.SetTZDatabase(time.Local)
	if err = DB.Sync2(new(Timer), new(Log), new(Setting)); err != nil {
		return err
	}

	data := []*Timer(nil)
	DB.Find(&data)
	for i := range data {
		v := data[i]
		if !v.Enable {
			continue
		}
		if v.Cron == Startup {
			ExecWithLog(v)
			continue
		}
		Corn.SetTask(conv.String(v.ID), v.Cron, func() { ExecWithLog(v) })
	}

	// 清理过期执行日志(保留近 7 天):启动时清理一次(适配开机自启) + 每天 3 点定时清理
	cleanupOldLogs()
	Corn.SetTask("cleanup_logs", "0 0 3 * * *", cleanupOldLogs)

	return nil
}

func Run(port int, filename string) error {
	// 监听: 优先使用配置端口,被占用时回退到系统自动分配的空闲端口
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	return RunWithListener(ln, filename)
}

// RunWithListener 在指定 listener 上启动服务,调用方负责创建 listener(可获知实际端口)。
func RunWithListener(ln net.Listener, filename string) error {
	if err := _init(filename); err != nil {
		ln.Close()
		return err
	}
	initAuth()
	s := fbr.Default()
	s.GET("/", func(c fbr.Ctx) {
		if sessionToken != "" && c.Cookies(sessionCookieName) != sessionToken {
			c.RedirectTo("/login")
			return
		}
		data, _ := webFS.ReadFile("web/index.html")
		c.Html200(data)
	})
	s.GET("/login", func(c fbr.Ctx) {
		data, _ := webFS.ReadFile("web/login.html")
		c.Html200(data)
	})

	// 读取嵌入的静态文件到 map,通过 /static/ 接口提供。
	staticSub, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return err
	}
	staticFiles := map[string][]byte{}
	fs.WalkDir(staticSub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		b, e := fs.ReadFile(staticSub, path)
		if e == nil {
			staticFiles["/"+path] = b
		}
		return nil
	})
	s.GET("/static/*", func(c fbr.Ctx) {
		// gofiber 中 "*" 为 wildcard 参数;不能用 c.Get("*")(那是取请求头)
		path := "/" + c.Params("*")
		if data, ok := staticFiles[path]; ok {
			// 根据后缀设置 Content-Type
			ct := "text/plain"
			switch {
			case len(path) > 3 && path[len(path)-3:] == ".js":
				ct = "application/javascript"
			case len(path) > 4 && path[len(path)-4:] == ".css":
				ct = "text/css"
			}
			c.Response().Header.SetContentType(ct)
			c.Response().SetBodyRaw(data)
		} else {
			c.Status(404)
		}
	})
	s.Group("/api", func(g fbr.Grouper) {
		g.Use(authMiddleware)
		g.POST("/login", loginHandler)
		g.POST("/logout", logoutHandler)
		g.GET("/timer/all", GetTimerAll)    //列表
		g.POST("/timer", PostTimer)         //新建
		g.PUT("/timer", PutTimer)           //修改
		g.DELETE("/timer", DelTimer)        //删除
		g.PUT("/timer/enable", EnableTimer) //启用/禁用
		g.POST("/timer/test", TestTimer)
		g.POST("/timer/exec", ExecTimer) //立即执行

		g.GET("/log/list", GetLogList)    //日志列表
		g.DELETE("/log", ClearLogHandler) //清除日志

		g.ALL("/notice/ws", NoticeWS) //通知-websocket

		g.GET("/setting/error_handler", GetErrorHandler)
		g.PUT("/setting/error_handler", PutErrorHandler)
		g.POST("/setting/error_handler/test", TestErrorHandler)
	})
	return s.RunListener(ln)
}

func GetTimerAll(c fbr.Ctx) {
	data := []*Timer(nil)
	err := DB.Find(&data)
	c.CheckErr(err)
	for _, v := range data {
		v.Resp(Corn.GetTask(conv.String(v.ID)))
	}
	c.Succ(data)
}

func PostTimer(c fbr.Ctx) {
	t := &Timer{}
	c.Parse(t)
	_, err := DB.Insert(t)
	c.CheckErr(err)
	if t.Cron != Startup {
		err = Corn.SetTask(conv.String(t.ID), t.Cron, func() {
			ExecWithLog(t)
		})
		if err != nil {
			// cron 注册失败,删除已插入的数据库记录,避免脏数据
			DB.ID(t.ID).Delete(new(Timer))
			c.CheckErr(err)
			return
		}
		if !t.Enable {
			Corn.DelTask(conv.String(t.ID))
		}
	}
	c.Succ(nil)
}

func PutTimer(c fbr.Ctx) {
	t := &Timer{}
	c.Parse(t)

	// 只更新 Name/Cron/Content/Type,不覆盖 Enable(前端编辑弹窗未传 enable)
	_, err := DB.ID(t.ID).Cols("Name", "Cron", "Content", "Type").Update(t)
	c.CheckErr(err)

	// 查出完整数据(含 Enable),决定是否重新注册定时任务
	DB.ID(t.ID).Get(t)
	Corn.DelTask(conv.String(t.ID))
	if t.Enable && t.Cron != Startup {
		err = Corn.SetTask(conv.String(t.ID), t.Cron, func() {
			ExecWithLog(t)
		})
		c.CheckErr(err)
	}
	c.Succ(nil)
}

func DelTimer(c fbr.Ctx) {
	t := new(Timer)
	c.Parse(t)
	_, err := DB.ID(t.ID).Delete(new(Timer))
	c.CheckErr(err)
	Corn.DelTask(conv.String(t.ID))
	c.Succ(nil)
}

func EnableTimer(c fbr.Ctx) {
	t := new(Timer)
	c.Parse(t)

	// 查出完整的任务数据(含 Cron/Content),前端只传了 id 和 enable
	full := new(Timer)
	if _, err := DB.ID(t.ID).Get(full); err != nil {
		c.CheckErr(err)
		return
	}

	err := DB.SessionFunc(func(session *xorm.Session) error {
		if _, err := session.ID(t.ID).Cols("Enable").Update(t); err != nil {
			return err
		}
		if t.Enable && full.Cron != Startup {
			if err := Corn.SetTask(conv.String(t.ID), full.Cron, func() {
				ExecWithLog(full)
			}); err != nil {
				return err
			}
		} else {
			Corn.DelTask(conv.String(t.ID))
		}
		return nil
	})
	c.CheckErr(err)
	c.Succ(nil)
}

func TestTimer(c fbr.Ctx) {
	t := &Timer{}
	c.Parse(t)
	result, err := Exec(t)
	// 记录测试日志
	l := &Log{TimerID: t.ID, Name: t.Name, Type: t.Type}
	if err != nil {
		l.Status = "error"
		l.Error = err.Error()
		logs.Errorf("脚本测试错误[%s]: %v", t.Name, err)
	} else {
		l.Status = "success"
		l.Result = conv.String(result)
		if len(l.Result) > 2000 {
			l.Result = l.Result[:2000] + "..."
		}
	}
	DB.Insert(l)
	if err != nil {
		c.JSON(map[string]interface{}{"code": 500, "msg": err.Error()})
		return
	}
	// 返回执行结果
	resultStr := conv.String(result)
	if len(resultStr) > 2000 {
		resultStr = resultStr[:2000] + "..."
	}
	c.JSON(map[string]interface{}{"code": 200, "data": resultStr})
}

// ExecTimer 立即执行一次已保存的任务(按 id 加载,异步执行)。
// 结果通过 WebSocket 通知 + 写入执行日志,接口立即返回。
func ExecTimer(c fbr.Ctx) {
	t := &Timer{}
	c.Parse(t)
	full := new(Timer)
	has, err := DB.ID(t.ID).Get(full)
	c.CheckErr(err)
	if !has {
		c.JSON(map[string]interface{}{"code": 404, "msg": "任务不存在"})
		return
	}
	go ExecWithLog(full)
	c.Succ(nil)
}

func NoticeWS(c fbr.Ctx) {
	c.Websocket(func(conn *fbr.Websocket) {
		client := &wsClient{conn: conn}
		WS.Set(conn, client)
		defer WS.Del(conn)
		<-conn.Done()
	})
}

func GetLogList(c fbr.Ctx) {
	timerID := conv.Int64(c.Query("timerId"))
	limit := conv.Int(c.Query("limit"))
	data, err := GetLogs(timerID, limit)
	c.CheckErr(err)
	c.Succ(data)
}

func ClearLogHandler(c fbr.Ctx) {
	timerID := conv.Int64(c.Query("timerId"))
	err := ClearLogs(timerID)
	c.CheckErr(err)
	c.Succ(nil)
}

// GetErrorHandler 获取错误处理脚本内容
func GetErrorHandler(c fbr.Ctx) {
	c.Succ(getErrorHandlerScript())
}

// PutErrorHandler 保存错误处理脚本(保存即生效; 传空脚本即停用错误处理)
func PutErrorHandler(c fbr.Ctx) {
	req := struct {
		Script string `json:"script"`
	}{}
	c.Parse(&req)
	if err := setErrorHandlerScript(req.Script); err != nil {
		c.CheckErr(err)
		return
	}
	c.Succ(nil)
}

// TestErrorHandler 用模拟参数测试错误处理脚本
func TestErrorHandler(c fbr.Ctx) {
	req := struct {
		Script string `json:"script"`
	}{}
	c.Parse(&req)
	code := req.Script
	if code == "" {
		code = getErrorHandlerScript()
	}
	if code == "" {
		c.JSON(map[string]interface{}{"code": 400, "msg": "脚本内容为空"})
		return
	}
	result, err := Script.CallFunc(code, "OnError", "999", "测试任务", "测试错误信息")
	if err != nil {
		c.JSON(map[string]interface{}{"code": 500, "msg": err.Error()})
		return
	}
	data := conv.String(result)
	if len(data) > 2000 {
		data = data[:2000] + "..."
	}
	c.JSON(map[string]interface{}{"code": 200, "data": data})
}
