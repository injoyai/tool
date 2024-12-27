package main

import (
	"github.com/injoyai/goutil/frame/in/v3"
	"github.com/injoyai/goutil/frame/mux"
	"io"
	"os"
)

func main() {

	s := mux.New()
	s.ALL("/video", func(r *mux.Request) {
		ws := r.Websocket()
		defer ws.Close()
		f, err := os.Open("F:\\test\\x36xhzz.mp4")
		in.CheckErr(err)
		ws.DiscardRead()
		io.Copy(ws, f)
	})

	s.SetPort(8090)
	s.Run()
}
