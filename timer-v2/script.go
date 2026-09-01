package timer

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs/v2"
	"github.com/injoyai/notice/pkg/push"
	"github.com/injoyai/notice/pkg/push/serverchan"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tool/timer/lib"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var (
	_tdx *tdx.Client
	_map = maps.NewSafe()
)

func init() {
	var err error
	_tdx, err = tdx.DialDefault()
	logs.PrintErr(err)
}

// scriptEngine 基于 yaegi 的 Go 脚本引擎,替代旧的 otto(js) 池。
//
// 脚本内容为标准 Go 代码,可使用标准库,以及通过 SetFunc 注册到 "i" 包的自定义函数。
// 示例脚本:
//
//	package main
//
//	import (
//		"fmt"
//		"time"
//		"i"
//	)
//
//	func main() {
//		i.Start("notepad")
//		r, _ := i.Ping("127.0.0.1", time.Second)
//		fmt.Println(r)
//	}
type scriptEngine struct {
	mu      sync.RWMutex
	symbols interp.Exports
}

func newScriptEngine() *scriptEngine {
	s := &scriptEngine{
		symbols: interp.Exports{
			"i/i": make(map[string]reflect.Value),
		},
	}
	s.registerBuiltins()
	return s
}

// SetFunc 将自定义函数注册到脚本可调用的 "i" 包下。
// fn 必须是函数类型,参数和返回值应使用 yaegi 可识别的基础类型
// (如 string/int/time.Duration/error 等)。
func (s *scriptEngine) SetFunc(name string, fn interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbols["i/i"][name] = reflect.ValueOf(fn)
}

// Exec 编译并执行一段 Go 脚本。
//
// 脚本应为完整的 Go 程序(包含 package main 与 func main)。
// 每次执行独立创建解释器,互不影响。
func (s *scriptEngine) Exec(code string) (interface{}, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, nil
	}
	var buf bytes.Buffer
	i := interp.New(interp.Options{
		Stdout: &buf,
		Stderr: &buf,
	})
	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, err
	}
	if err := i.Use(lib.Symbols); err != nil {
		return nil, err
	}
	s.mu.RLock()
	err := i.Use(s.symbols)
	s.mu.RUnlock()
	if err != nil {
		return nil, err
	}
	v, err := i.Eval(code)
	output := buf.String()
	if err != nil {
		if output != "" {
			return output, err
		}
		return nil, err
	}
	if output != "" {
		return output, nil
	}
	if v.IsValid() {
		return v.Interface(), nil
	}
	return nil, nil
}

// CallFunc 执行脚本并调用其中名为 fn 的函数, args 均以 string 传入。
// 用于错误处理: 先 Eval 完整脚本完成声明(注意: 若脚本含 func main 会在此被执行,
// 错误处理脚本应只包含 OnError 等函数声明), 再 Eval "fn(int64(id), ...)" 完成带参调用。
// args[0] 必须是十进制数字字符串, 转为 int64 传参; 其余按 string 传参。
func (s *scriptEngine) CallFunc(code, fn string, args ...string) (interface{}, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("脚本内容为空")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("缺少函数参数")
	}

	// 构造调用表达式: fn(int64(123), "name", "msg")
	// args[0] 必须是十进制数字字符串(调用方用 FormatInt 生成), 不能加引号
	// (int64("123") 是非法 Go 表达式, 编译报错)
	callArgs := make([]string, 0, len(args))
	callArgs = append(callArgs, fmt.Sprintf("int64(%s)", args[0]))
	for _, a := range args[1:] {
		callArgs = append(callArgs, strconv.Quote(a))
	}
	expr := fmt.Sprintf("%s(%s)", fn, strings.Join(callArgs, ", "))

	var buf bytes.Buffer
	i := interp.New(interp.Options{
		Stdout: &buf,
		Stderr: &buf,
	})
	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, err
	}
	if err := i.Use(lib.Symbols); err != nil {
		return nil, err
	}
	s.mu.RLock()
	err := i.Use(s.symbols)
	s.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	if _, err := i.Eval(code); err != nil {
		if output := buf.String(); output != "" {
			return output, err
		}
		return nil, err
	}
	v, err := i.Eval(expr)
	if err != nil {
		if output := buf.String(); output != "" {
			return output, err
		}
		return nil, err
	}
	if buf.String() != "" {
		return buf.String(), nil
	}
	if v.IsValid() {
		return v.Interface(), nil
	}
	return nil, nil
}

// registerBuiltins 注册定时任务常用的内置函数。
func (s *scriptEngine) registerBuiltins() {

	s.SetFunc("Set", _map.Set)
	s.SetFunc("Get", _map.Get)
	s.SetFunc("Del", _map.Del)

	// Start 启动一个外部程序(或打开文件/网址)。
	s.SetFunc("Start", func(cmd string) error {
		return shell.Start2(cmd)
	})

	// Ping 探测主机连通性,返回结果描述。
	s.SetFunc("Ping", func(host string, timeout time.Duration) (string, error) {
		r, err := ip.Ping(host, timeout)
		if err != nil {
			return "", err
		}
		return r.String(), nil
	})

	// Notice 发送通知。
	s.SetFunc("Notice", Notice)

	// Dial 探测网络端口连通性。
	s.SetFunc("Dial", func(network, address string, timeout time.Duration) (string, error) {
		c, err := net.DialTimeout(network, address, timeout)
		if err != nil {
			return "", err
		}
		c.Close()
		return "成功", nil
	})

	// DialTCP 探测 TCP 端口连通性。
	s.SetFunc("DialTCP", func(address string, timeout time.Duration) (string, error) {
		c, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return "", err
		}
		c.Close()
		return "成功", nil
	})

	// ServerChan 发送消息到 Server 频道。
	s.SetFunc("ServerChan", func(title, msg string) error {
		key := cfg.GetString("notice.serverchan.key")
		if key == "" {
			return fmt.Errorf("notice.serverchan.key is empty")
		}
		return serverchan.New(key).Push(&push.Message{
			Title:   title,
			Content: msg,
		})
	})

	// QuarkCheckin 夸克签到
	s.SetFunc("QuarkCheckin", func(vcode, sign, kps string) (string, error) {

		debug := false
		queries := map[string]any{
			"fr":    "android",
			"pr":    "ucpro",
			"vcode": vcode,
			"sign":  sign,
			"kps":   kps,
		}

		infoUrl := "https://drive-m.quark.cn/1/clouddrive/capacity/growth/info"
		resp := http.Url(infoUrl).SetQuerys(queries).Debug(debug).Get()
		if resp.Err() != nil {
			return "", resp.Err()
		}

		info := new(quarkInfoResp)
		err := resp.Bind(info)
		if err != nil {
			return "", err
		}
		if info.Status == 200 && info.Data.CapSign.SignDaily {
			return info.String(), nil
		}

		signUrl := "https://drive-m.quark.cn/1/clouddrive/capacity/growth/sign"
		resp = http.Url(signUrl).SetQuerys(queries).Debug(debug).Post()
		if resp.Err() != nil {
			return "", resp.Err()
		}

		resp = http.Url(infoUrl).SetQuerys(queries).Debug(debug).Get()
		if resp.Err() != nil {
			return "", resp.Err()
		}

		info = new(quarkInfoResp)
		_ = resp.Bind(info)

		return info.String(), nil
	})

	s.SetFunc("GetPrices", func(code ...string) (map[string]float64, error) {
		resp, err := _tdx.GetQuote(code...)
		if err != nil {
			return nil, err
		}
		if len(resp) == 0 {
			return nil, fmt.Errorf("未找到%s", code)
		}
		m := map[string]float64{}
		for _, v := range resp {
			m[v.Exchange.String()+v.Code] = v.Kline.Close.Float64()
		}
		now := time.Now()
		now.Weekday()
		return m, nil
	})

}

type quarkInfoResp struct {
	Status int `json:"status"`
	Data   struct {
		CapSign struct {
			SignDaily       bool  `json:"sign_daily"`        //是否签到
			SignProgress    int   `json:"sign_progress"`     //签到进度
			SignTarget      int   `json:"sign_target"`       //签到目标
			SignDailyReward int64 `json:"sign_daily_reward"` //每日签到奖励
		} `json:"cap_sign"`
		CapGrowth struct {
			CurTotalCap int64 `json:"cur_total_cap"` //签到累计奖励
		} `json:"cap_growth"`
	} `json:"data"`
}

func (this *quarkInfoResp) String() string {
	return fmt.Sprintf("签到进度:%d/%d, 签到奖励:%s, 累计签到:%s",
		this.Data.CapSign.SignProgress, this.Data.CapSign.SignTarget,
		oss.SizeString(this.Data.CapSign.SignDailyReward),
		oss.SizeString(this.Data.CapGrowth.CurTotalCap))
}
