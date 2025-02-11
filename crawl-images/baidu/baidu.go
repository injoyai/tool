package baidu

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	Url         = `https://image.baidu.com/search/acjson?tn=resultjson_com&logid=11427217667241829936&ipn=rj&ct=201326592&is=&fp=result&fr=&word=%s&queryWord=%s&cl=2&lm=&ie=utf-8&oe=utf-8&adpicid=&st=-1&z=&ic=0&hd=&latest=&copyright=&s=&se=&tab=&width=&height=&face=0&istype=2&qc=&nc=1&expermode=&nojc=&isAsync=&pn=%d&rn=%d&gsm=1e&1739240627486=`
	DefaultSize = 30
	UserAgent   = http.UserAgentDefault
)

func Crawl(key string, dir string, limit int) error {

	key = url.QueryEscape(key)
	os.MkdirAll(dir, os.ModePerm)

	for i := DefaultSize; i <= limit; i += DefaultSize {

		u := fmt.Sprintf(Url, key, key, i, DefaultSize)

		bs, err := http.Url(u).SetHeaders(map[string][]string{
			"User-Agent":      {UserAgent},
			"Cookie":          {"BAIDUID=20DE567559ACCE6966E08CFC53F3C6D1:FG=1; BAIDUID_BFESS=1493866524C32CF5F689BC5CF753F0D8:FG=1; BDRCVFR[-pGxjrCMryR]=mk3SLVN4HKm; BIDUPSID=1493866524C32CF5DBCA1FA6598C4DDB; H_PS_PSSID=60819_60843; H_WISE_SIDS=61803_62111; PSTM=1701757316"},
			"Connection":      {"keep-alive"},
			"Content-Type":    {"application/json"},
			"accept":          {"text/plain, */*; q=0.01"},
			"accept-language": {"zh-CN,zh;q=0.9"},
			"referer":         {"https://image.baidu.com/search/index"},
		}).Debug(false).GetBytes()
		if err != nil {
			logs.Err(err)
			continue
		}

		for _, v := range conv.NewMap(bs).GetInterfaces("data") {
			imagesUrl := conv.NewMap(v).GetString("replaceUrl[0].ObjURL")
			if imagesUrl == "" {
				continue
			}
			uu, err := url.Parse(imagesUrl)
			if err != nil {
				logs.Err(err)
				continue
			}
			filename := filepath.Join(dir, path.Base(uu.Path))
			if filename == "" {
				continue
			}
			switch {
			case strings.HasSuffix(filename, ".jpg"):
			case strings.HasSuffix(filename, ".jpeg"):
			case strings.HasSuffix(filename, ".png"):
			default:
				filename += ".jpg"
			}
			logs.Debug(imagesUrl, filename)

			bs, err := http.GetBytes(imagesUrl)
			if err != nil {
				//logs.Err(err)
				continue
			}
			oss.New(filename, bs)
		}

	}

	return nil
}
