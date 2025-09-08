package main

import (
	_ "embed"
	"github.com/injoyai/lorca"
	"log"
)

//go:embed compound_interest.html
var html string

func main() {
	err := lorca.Run(&lorca.Config{
		Width:  660,
		Height: 820,
		Index:  html,
	})
	log.Println(err)
}
