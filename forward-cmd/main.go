package main

import (
	"context"
	"github.com/injoyai/goutil/other/command"
	"github.com/injoyai/logs"
	"github.com/injoyai/proxy/core"
	"github.com/injoyai/proxy/forward"
	"github.com/spf13/cobra"
)

func main() {

	c := &command.Command{
		Command: cobra.Command{
			Use:     "forward",
			Short:   "端口转发",
			Example: "forward -p=8000 -a=127.0.0.1:8001",
		},
		Flag: []*command.Flag{
			{Name: "port", Short: "p", Memo: "端口", Default: "8080"},
			{Name: "address", Short: "a", Memo: "转发地址", Default: ":8081"},
		},
		Run: func(cmd *cobra.Command, args []string, flag *command.Flags) {
			f := &forward.Forward{
				Listen:  core.NewListenTCP(flag.GetInt("port")),
				Forward: core.NewDialTCP(flag.GetString("address")),
			}
			err := f.Run(context.Background())
			logs.Err(err)
		},
	}

	c.Execute()

}
