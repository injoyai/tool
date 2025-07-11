package upgrade

import (
	"errors"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"os"
	"time"
)

var Log = logs.New("Upgrade").SetFormatter(logs.TimeFormatter)

type Upgrade struct {
	Filename string       //文件名
	Retry    int          //重试次数
	Backups  string       //备份路径("./bak/{name}(20060102)"),空则不备份
	Stop     func() error //杀死进程命令
	Start    func() error //启动进程命令
}

func (this *Upgrade) Upgrade(info *Info) error {
	return g.Retry(func() error {
		return this.upgrade(info)
	}, this.Retry, time.Second*2)
}

func (this *Upgrade) upgrade(info *Info) (err error) {

	defer func() {
		if err != nil {
			Log.Printf("升级失败: %v\n\n", err)
			return
		}
		Log.Println("升级成功", info.Version)
	}()

	temp := this.Filename + ".downloading"
	defer os.Remove(temp)

	//下载升级文件
	Log.Printf("下载升级文件: %s  代理: %s\n", info.Url, info.Proxy)
	result, err := Download(info.Url, temp, info.Proxy)
	if err != nil {
		return err
	}

	//判断文件是否正确
	if len(info.MD5) > 0 && result.MD5 != info.MD5 {
		return errors.New("文件MD5校验失败")
	}

	//关闭旧程序
	if this.Stop != nil {
		if err := this.Stop(); err != nil {
			return err
		}
	}

	//备份旧版本
	if len(this.Backups) > 0 && oss.Exists(this.Filename) {
		bakName := time.Now().Format(this.Backups)
		if err := os.Rename(this.Filename, bakName); err != nil {
			return err
		}
	}

	//替换新版本
	if err := os.Rename(temp, this.Filename); err != nil {
		return err
	}

	//启动新程序
	if this.Start != nil {
		if err := this.Start(); err != nil {
			return err
		}
	}

	return nil
}

type Info struct {
	Url     string //下载地址
	Proxy   string //代理
	Version string //版本
	MD5     string //md5,为空则不校验
}
