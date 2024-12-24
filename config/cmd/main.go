package main

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/goutil/other/command"
	"github.com/injoyai/tool/config"
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
			{Name: "natures", Short: "n", Memo: "配置属性路径(json)", Default: "./config/nature.json"},
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
			if err := config.GUI(config.New(filename, natures)); err != nil {
				fmt.Println(err)
			}
		},
	}
	root.Execute()
}

var ExampleNatures = []*config.Nature{
	{Key: "nickName", Name: "昵称"},
	{Key: "resource", Name: "资源地址"},
	{Key: "proxy", Name: "默认代理地址"},
	{Key: "proxyIgnore", Name: "忽略代理正则"},
	{Key: "memoHost", Name: "备注请求地址"},
	{Key: "memoToken", Name: "备注API秘钥"},
	{Key: "uploadMinio", Name: "Minio上传配置", Type: "object2", Value: []config.Nature{
		{Name: "请求地址", Key: "endpoint"},
		{Name: "AccessKey", Key: "access"},
		{Name: "SecretKey", Key: "secret"},
		{Name: "存储桶", Key: "bucket"},
		{Name: "随机名称", Key: "rename", Type: "bool"},
	}},
	{Key: "downloadDir", Name: "默认下载地址"},
	{Key: "downloadNoticeEnable", Name: "默认启用通知", Type: "bool"},
	{Key: "downloadNoticeText", Name: "默认通知内容"},
	{Key: "downloadVoiceEnable", Name: "默认启用语音", Type: "bool"},
	{Key: "downloadVoiceText", Name: "默认语音内容"},
	{Key: "customOpen", Name: "自定义打开", Type: "object"},
}
