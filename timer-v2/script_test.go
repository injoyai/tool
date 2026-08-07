package timer

import (
	"strings"
	"testing"
)

// TestScriptExec_Stdlib 验证标准库可用。
func TestScriptExec_Stdlib(t *testing.T) {
	s := newScriptEngine()
	code := `package main

import "fmt"

func main() {
	fmt.Println("hello yaegi")
}
`
	if _, err := s.Exec(code); err != nil {
		t.Fatalf("执行标准库脚本失败: %v", err)
	}
}

// TestScriptExec_Funcs 验证自定义 funcs 包函数能被脚本调用。
func TestScriptExec_Funcs(t *testing.T) {
	s := newScriptEngine()
	s.SetFunc("Echo", func(x string) string { return "echo:" + x })

	code := `package main

import (
	"fmt"
	"i"
)

func main() {
	fmt.Println(i.Echo("hi"))
}
`
	if _, err := s.Exec(code); err != nil {
		t.Fatalf("执行自定义函数脚本失败: %v", err)
	}
}

// TestScriptExec_Variadic 验证 variadic 参数的自定义函数可用。
func TestScriptExec_Variadic(t *testing.T) {
	s := newScriptEngine()
	got := ""
	s.SetFunc("Join", func(parts ...string) string {
		got = strings.Join(parts, "-")
		return got
	})
	code := `package main

import (
	"fmt"
	"i"
)

func main() {
	fmt.Println(i.Join("a", "b", "c"))
}
`
	if _, err := s.Exec(code); err != nil {
		t.Fatalf("执行 variadic 脚本失败: %v", err)
	}
	if got != "a-b-c" {
		t.Fatalf("variadic 传参异常, got=%q", got)
	}
}

// TestScriptExec_Empty 空脚本应直接返回。
func TestScriptExec_Empty(t *testing.T) {
	s := newScriptEngine()
	if _, err := s.Exec("   "); err != nil {
		t.Fatalf("空脚本不应报错: %v", err)
	}
}

// TestScriptExec_SyntaxError 语法错误应返回 error 而非 panic。
func TestScriptExec_SyntaxError(t *testing.T) {
	s := newScriptEngine()
	if _, err := s.Exec("package main\nfunc main( { }"); err == nil {
		t.Fatal("期望语法错误,但未报错")
	}
}

// TestScriptExec_LibSymbols 验证 lib 符号(conv/logs/crypt 等)在脚本中可用。
func TestScriptExec_LibSymbols(t *testing.T) {
	s := newScriptEngine()
	code := `package main

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/logs/v2"
)

func main() {
	s := conv.String(42)
	fmt.Println(s)
	logs.Info("test from lib symbols")
}
`
	if _, err := s.Exec(code); err != nil {
		t.Fatalf("执行 lib 符号脚本失败: %v", err)
	}
}
