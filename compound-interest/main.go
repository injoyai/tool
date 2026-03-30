package main

import (
	_ "embed"
	"log"

	"github.com/injoyai/lorca"
)

//go:embed compound_interest.html
var html string

func main() {
	err := lorca.Run(&lorca.Config{
		Width:  800,
		Height: 890,
		Index:  html,
	})
	log.Println(err)
}
