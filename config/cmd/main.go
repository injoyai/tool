package main

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/goutil/other/command"
	"github.com/injoyai/tool/config"
	"github.com/injoyai/tool/config/testdata"
	"github.com/spf13/cobra"
	"os"
)

func main() {

	root := command.Command{
		Command: cobra.Command{
			Use:   "config",
			Short: "config -f [filename] -n [natures]",
		},
		Flag: []*command.Flag{
			{Name: "filename", Short: "f", Memo: "配置文件路径", Default: "./config/config.yaml"},
			{Name: "natures", Short: "n", Memo: "配置属性路径(json)", Default: "./config/config_nature.json"},
		},
		Run: func(cmd *cobra.Command, args []string, flags *command.Flags) {
			filename := flags.GetString("filename")
			natures := config.Natures(nil)
			naturesFile := flags.GetString("natures")
			bs, err := os.ReadFile(naturesFile)
			if err != nil {
				fmt.Println(err)
				return
			}
			if err := json.Unmarshal(bs, &natures); err != nil {
				fmt.Println(err)
				return
			}
			natures = testdata.ExampleNatures
			if err := config.GUI(config.New(filename, natures)); err != nil {
				fmt.Println(err)
			}
		},
	}
	root.Execute()
}
