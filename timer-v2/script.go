package timer

import (
	"bytes"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/tool/timer/lib"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

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

// registerBuiltins 注册定时任务常用的内置函数。
func (s *scriptEngine) registerBuiltins() {
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
	s.SetFunc("Notice", func(msg string) error {
		return notice.DefaultWindows.Publish(&notice.Message{
			Content: msg,
		})
	})

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

}
