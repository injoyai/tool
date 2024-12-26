package main

import (
	"context"
	"github.com/injoyai/proxy/forward"
)

func Run(f *forward.Forward) { f.Run(context.Background()) }
