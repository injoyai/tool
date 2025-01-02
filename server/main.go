package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-toast/toast"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/frame/in/v3"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/oss/tray"
	"github.com/injoyai/goutil/types"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/frame"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"github.com/injoyai/tool/config"
	"github.com/injoyai/tool/server/file"
	"io"
	"net"
	"net/http"
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

	rerun := chans.NewRerun(Server)
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
				s.AddMenu().SetName(k).OnClick(func(m *tray.Menu) {
					shell.Run(conv.String(v))
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

func Server(ctx context.Context) {
	cfg.Init(cfg.WithFile(Filename))
	go httpServer(ctx, cfg.GetInt("http_port"))
	tcpServer(ctx, cfg.GetInt("tcp_port"))
}

func httpServer(ctx context.Context, port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		l.Close()
	}()
	logs.Infof("[:%d] 开启HTTP服务成功...\n", port)
	return http.Serve(l, in.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		msg := &types.Message{}

		switch r.URL.Path {
		case "", "/":
			msg = &types.Message{
				Type: "shell",
				Data: r.URL.Query().Get("cmd"),
			}

		case "/file", "/files":
			files := []*file.File(nil)
			err := r.ParseMultipartForm(32 << 20)
			in.CheckErr(err)
			for filename, fs := range r.MultipartForm.File {
				for _, f := range fs {
					func() {
						fileReader, err := f.Open()
						in.CheckErr(err)
						defer fileReader.Close()
						bs, err := io.ReadAll(fileReader)
						in.CheckErr(err)
						files = append(files, &file.File{
							Filename: filename,
							Data:     base64.StdEncoding.EncodeToString(bs),
						})
					}()
					break
				}
			}
			msg.Data = files

		default:
			defer r.Body.Close()
			bs, err := io.ReadAll(r.Body)
			in.CheckErr(err)
			msg = &types.Message{
				Type: r.URL.Path[1:],
				Data: bs,
			}

		}

		err := deal(msg)
		in.CheckErr(err)

		in.Succ(msg.Data)

	})))
}

func tcpServer(ctx context.Context, port int) error {
	return listen.RunTCPContext(ctx, port, func(s *server.Server) {
		s.Logger.Debug(false)
		s.SetClientOption(func(c *client.Client) {
			c.Logger.Debug(false)
			c.Event.WithFrame(frame.Default)
			c.OnDealMessage = func(c *client.Client, msg ios.Acker) {

				m := &types.Message{}
				if err := json.Unmarshal(msg.Payload(), m); err != nil {
					logs.Err(err)
					return
				}

				err := deal(m)
				if err != nil {
					logs.Err(err)
					resp := m.Response(Fail, nil, err.Error())
					c.WriteAny(resp)
					return
				}

				resp := m.Response(Succ, nil, "成功")
				c.WriteAny(resp)

			}
		})
		logs.Infof("[:%d] 开启TCP服务成功...\n", port)
	})
}

func deal(msg *types.Message) (err error) {

	if msg == nil {
		return nil
	}

	if msg.IsResponse() {
		return nil
	}

	data := msg.Data
	msg.Data = nil

	switch msg.Type {

	case "ping":
		msg.Data = g.Map{
			"uptime":  g.Uptime.Unix(),
			"version": Version,
		}
		msg.Msg = "pong"

	case "shell":
		err = shell.Run(conv.String(data))

	case "deploy", "file", "files":
		err = file.Do(data)

		/*



		 */

	case "edge.notice.upgrade":
		m := conv.NewMap(data)

		//显示通知和是否升级按钮按钮
		upgradeEdge := fmt.Sprintf("http://localhost:%d", cfg.GetInt("http_port")) + "?cmd=in%20server%20edge%20upgrade"
		notification := toast.Notification{
			AppID:   "Microsoft.Windows.Shell.RunDialog",
			Title:   fmt.Sprintf("发现新版本(%s),是否马上升级?", m.GetString("version")),
			Message: "版本详情: " + m.GetString("versionDetails"),
			Actions: []toast.Action{
				{"protocol", "马上升级", upgradeEdge},
				{"protocol", "稍后再说", ""},
			},
		}
		if err := notification.Push(); err != nil {
			return err
		}

		//播放语音
		notice.DefaultVoice.Speak(fmt.Sprintf("主人. 发现网关新版本(%s). 是否马上升级?", m.GetString("version")))

	case "edge.upgrade":
		err = shell.Start("in server edge upgrade")

	case "edge.open", "edge.run", "edge.start":
		err = shell.Start("in server edge")

	case "edge.close", "edge.stop", "edge.shutdown":
		err = shell.Start("in server edge stop")

	default:
		err = errors.New("未知命令: " + msg.Type)

	}

	return
}
