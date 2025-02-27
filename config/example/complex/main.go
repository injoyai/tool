package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/tool/config"
	"log"
)

func main() {
	err := config.GUI(config.New("./example/complex/config.yaml", Natures).
		SetWidthHeight(720, 680).
		SetOnSaved(func(m *conv.Map) {
			log.Println("保存成功")
			fmt.Println(m.String())
		}))
	log.Println(err)
}

var Natures = []*config.Nature{
	{Key: "bool", Name: "布尔", Type: "bool"},
	{Key: "nickName", Name: "昵称"},
	{Key: "resource", Name: "资源地址"},
	{Key: "proxy", Name: "默认代理地址"},
	{Key: "proxyIgnore", Name: "忽略代理正则"},
	{Key: "memoHost", Name: "备注请求地址"},
	{Key: "memoToken", Name: "备注API秘钥"},
	{Key: "uploadMinio", Name: "Minio上传配置", Type: "object", Value: []config.Nature{
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
	{Key: "customOpen", Name: "自定义打开", Type: "map"},
}
