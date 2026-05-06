package service

import (
	"context"
	"time"

	"github.com/injoyai/base/maps"
	"github.com/injoyai/frame/fbr"
	"github.com/injoyai/logs"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

func (this *Server) TDX(ctx context.Context, port int) error {
	cli, err := tdx.DialDefault()
	logs.PanicErr(err)

	minuteCache := maps.NewGeneric[string, *protocol.MinuteResp]()

	s := fbr.Default(fbr.WithPort(port), fbr.WithContext(ctx))
	s.ALL("/minute/ws", func(c fbr.Ctx) {
		code := c.GetString("code")
		c.Websocket(func(ws *fbr.Websocket) {
			go ws.DiscardRead()
			t := time.NewTicker(time.Second)
			for range t.C {
				resp, err := minuteCache.GetOrSetByHandler(code, func() (*protocol.MinuteResp, error) {
					return cli.GetMinute(code)
				}, time.Second)
				if err != nil {
					logs.Error(err)
					continue
				}
				err = ws.WriteJSON(resp.List)
				if err != nil {
					return
				}
			}
		})
	})
	return s.Run()
}
