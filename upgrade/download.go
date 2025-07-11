package upgrade

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/str/bar/v2"
	"time"
)

// Download 下载升级文件
func Download(url string, filename string, proxy ...string) (*Result, error) {

	c := http.NewClient().SetTimeout(0)
	if err := c.SetProxy(conv.Default("", proxy...)); err != nil {
		return nil, err
	}

	now := time.Now()
	m := md5.New()
	b := bar.New()
	defer b.Close()
	size, err := c.GetToFileWithPlan(url, filename, func(p *http.Plan) {
		b.SetTotal(p.Total)
		b.Set(p.Current)
		b.Flush()
		m.Write(*p.Bytes)
	})
	if err != nil {
		return nil, err
	}
	info := &Result{
		Size:  size,
		MD5:   hex.EncodeToString(m.Sum(nil)),
		Spend: time.Since(now),
	}
	return info, nil
}

type Result struct {
	Size  int64
	MD5   string
	Spend time.Duration
}
