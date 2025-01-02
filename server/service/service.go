package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/frame/in/v3"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/types"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/frame"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"github.com/injoyai/tool/server/edge"
	"github.com/injoyai/tool/server/file"
	"io"
	"net"
	"net/http"
	"strings"
)

type Server struct {
	Filename string
	Version  string
	Fail     int
	Succ     int
}

func (this *Server) Run(ctx context.Context) {
	cfg.Init(cfg.WithFile(this.Filename))
	go this.HTTP(ctx, cfg.GetInt("http_port"))
	this.TCP(ctx, cfg.GetInt("tcp_port"))
}

func (this *Server) HTTP(ctx context.Context, port int) error {
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

		err := this.deal(msg)
		in.CheckErr(err)

		in.Succ(msg.Data)

	})))
}

func (this *Server) TCP(ctx context.Context, port int) error {
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

				err := this.deal(m)
				if err != nil {
					logs.Err(err)
					resp := m.Response(this.Fail, nil, err.Error())
					c.WriteAny(resp)
					return
				}

				resp := m.Response(this.Succ, nil, "成功")
				c.WriteAny(resp)

			}
		})
		logs.Infof("[:%d] 开启TCP服务成功...\n", port)
	})
}

func (this *Server) deal(msg *types.Message) (err error) {

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
			"version": this.Version,
		}
		msg.Msg = "pong"

	case "shell":
		err = shell.Run(conv.String(data))

	case "deploy", "file", "files":
		err = file.Do(data)

	/*



	 */

	default:

		switch {
		case strings.HasPrefix(msg.Type, "edge."):
			err = edge.Do(msg.Type, data)

		default:
			err = errors.New("未知命令: " + msg.Type)

		}

	}

	return
}
