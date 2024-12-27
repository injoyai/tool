package main

import (
	_ "embed"
	"github.com/injoyai/lorca"
)

//go:embed index.html
var index string

func main() {
	lorca.Run(&lorca.Config{
		Width:  860,
		Height: 690,
		Index:  index,
	})
}
