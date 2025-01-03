package file

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	Filename string `json:"filename"` //文件路径
	Data     []byte `json:"data"`     //文件内容
	Restart  bool   `json:"restart"`  //是否重启,针对于可执行文件
}

func Do(a any) error {
	var files []*File
	switch v := a.(type) {
	case []*File:
		files = v
	default:
		bs := conv.Bytes(a)
		err := json.Unmarshal(bs, &files)
		if err != nil {
			return err
		}
	}

	for _, v := range files {
		if v.Restart {
			logs.Info("关闭文件:", v.Filename)
			shell.Stop(v.Filename)
		}

		dir, _ := filepath.Split(v.Filename)

		zipPath := filepath.Join(dir, time.Now().Format("20060102150405.zip"))
		logs.Info("保存文件:", zipPath)
		if err := oss.New(zipPath, v.Data); err != nil {
			return fmt.Errorf("保存文件(%s)错误: %s", zipPath, err)
		}

		logs.Info("解压文件:", dir)
		if err := zip.Decode(zipPath, dir); err != nil {
			return fmt.Errorf("解压文件(%s)到(%s)错误: %s", zipPath, dir, err)
		}

		logs.Info("删除压缩包:", dir)
		os.Remove(zipPath)

		if v.Restart {
			logs.Info("执行文件:", v.Filename)
			if err := shell.Start(v.Filename); err != nil {
				return fmt.Errorf("执行文件(%s)错误: %s", v.Filename, err)
			}
		}
	}

	return nil
}
