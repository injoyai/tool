package main

import (
	_ "embed"
	"github.com/injoyai/lorca"
	"log"
	"net"
	"net/http"
)

//go:embed index.html
var index []byte

//go:embed hls.js
var hls []byte

func main() {
	err := lorca.Run(&lorca.Config{
		Width:  860,
		Height: 690,
		Index:  "",
	}, func(app lorca.APP) error {

		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}
		go http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/index.html", "/":
				w.Write(index)
			case "/hls.js":
				w.Write(hls)
			}
		}))
		return app.Load("http://" + l.Addr().String() + "/index.html")
	})
	log.Println(err)
}
