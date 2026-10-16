package timer

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/injoyai/logs/v2"
)

// ErrorHandlerScriptKey 错误处理脚本在 Setting 表中的 key
const ErrorHandlerScriptKey = "error_handler_script"

var (
	_errorHandlerMu       sync.RWMutex
	_errorHandlerCache    string
	_errorHandlerCacheSet bool
)

// getErrorHandlerScript 读取错误处理脚本(带进程内缓存)。
// 缓存未命中时查库, 查库失败记日志并返回缓存值(可能为空)。
func getErrorHandlerScript() string {
	// DB 为 nil(单测环境)时直接返回空串, 避免 Where 触发空指针 panic
	if DB == nil {
		return ""
	}
	_errorHandlerMu.RLock()
	if _errorHandlerCacheSet {
		v := _errorHandlerCache
		_errorHandlerMu.RUnlock()
		return v
	}
	_errorHandlerMu.RUnlock()

	_errorHandlerMu.Lock()
	defer _errorHandlerMu.Unlock()
	if _errorHandlerCacheSet { // double check
		return _errorHandlerCache
	}
	s := new(Setting)
	if has, err := DB.Where("`Key` = ?", ErrorHandlerScriptKey).Get(s); err != nil {
		logs.Errorf("读取错误处理脚本失败: %v", err)
	} else if has {
		_errorHandlerCache = s.Value
	}
	_errorHandlerCacheSet = true
	return _errorHandlerCache
}

// setErrorHandlerScript 保存错误处理脚本并刷新缓存。
func setErrorHandlerScript(v string) error {
	if has, err := DB.Where("`Key` = ?", ErrorHandlerScriptKey).Get(new(Setting)); err != nil {
		return err
	} else if has {
		if _, err := DB.Where("`Key` = ?", ErrorHandlerScriptKey).Cols("Value").Update(&Setting{Value: v}); err != nil {
			return err
		}
	} else {
		if _, err := DB.Insert(&Setting{Key: ErrorHandlerScriptKey, Value: v}); err != nil {
			return err
		}
	}
	_errorHandlerMu.Lock()
	_errorHandlerCache = v
	_errorHandlerCacheSet = true
	_errorHandlerMu.Unlock()
	return nil
}

// onError 任务执行失败时的全局处理入口: 加载错误处理脚本并调用 OnError(taskID, taskName, errMsg)。
// 脚本为空则静默跳过; 处理脚本自身失败仅记录(Type=error_handler), 绝不递归触发。
func onError(taskID int64, taskName, errMsg string) {
	defer func() {
		if r := recover(); r != nil {
			logs.Errorf("错误处理脚本 panic: %v", r)
		}
	}()

	// DB 为 nil(单测环境)时直接跳过
	if DB == nil {
		return
	}

	code := getErrorHandlerScript()
	if code == "" {
		return
	}

	result, err := Script.CallFunc(code, "OnError",
		strconv.FormatInt(taskID, 10), taskName, errMsg)

	l := &Log{
		TimerID: taskID,
		Name:    taskName,
		Type:    "error_handler",
	}
	if err != nil {
		l.Status = "error"
		l.Error = err.Error()
		logs.Errorf("错误处理脚本执行失败[%s]: %v", taskName, err)
	} else {
		l.Status = "success"
		l.Result = fmt.Sprint(result)
		if len(l.Result) > 2000 {
			l.Result = l.Result[:2000] + "..."
		}
	}
	if _, err := DB.Insert(l); err != nil {
		logs.Errorf("写入错误处理日志失败: %v", err)
	}
}
