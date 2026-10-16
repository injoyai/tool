package timer

import (
	"fmt"
	"time"

	"github.com/injoyai/conv"
	"github.com/injoyai/frame/fbr"
	"github.com/injoyai/logs/v2"
)

// Log 执行日志
type Log struct {
	ID        int64  `json:"id"`
	TimerID   int64  `json:"timerId"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"` // success / error
	Result    string `json:"result"`
	Error     string `json:"error"`
	CreatedAt string `json:"createdAt" xorm:"created"`
}

// ExecWithLog 执行任务并记录日志
func ExecWithLog(t *Timer) {
	log := &Log{
		TimerID: t.ID,
		Name:    t.Name,
		Type:    t.Type,
	}

	result, err := Exec(t)
	if err != nil {
		log.Status = "error"
		log.Error = err.Error()
		logs.Errorf("[%s] %v", t.Name, err)
		t.ExecErr = err.Error()
	} else {
		log.Status = "success"
		t.ExecErr = ""
		// 记录输出结果(截断过长的输出)
		log.Result = conv.String(result)
		if len(log.Result) > 2000 {
			log.Result = log.Result[:2000] + "..."
		}
	}

	// 写入日志
	if _, err := DB.Insert(log); err != nil {
		logs.Errorf("写入执行日志失败: %v", err)
	}

	// 执行失败时触发全局错误处理脚本(异步, 带任务ID/名称/错误信息)
	if log.Status == "error" {
		go onError(t.ID, t.Name, err.Error())
	}

	// 通过 WebSocket 推送通知
	noticeWS(log.Status == "success", t.Name, log.Error)
}

// GetLogs 查询日志,按 timerId 过滤,返回最近 limit 条
func GetLogs(timerID int64, limit int) ([]*Log, error) {
	logs := []*Log(nil)
	session := DB.Desc("id")
	if limit > 0 {
		session.Limit(limit)
	} else {
		session.Limit(200) // 默认最多 200 条
	}
	if timerID > 0 {
		session.Where("TimerID = ?", timerID)
	}
	err := session.Find(&logs)
	return logs, err
}

// ClearLogs 清除日志,timerID<=0 时清除全部
func ClearLogs(timerID int64) error {
	if timerID > 0 {
		_, err := DB.Where("TimerID = ?", timerID).Delete(new(Log))
		return err
	}
	_, err := DB.Exec("DELETE FROM log")
	return err
}

// cleanupOldLogs 删除超过 7 天的执行日志,避免日志堆积影响查询性能。
// CreatedAt 为字符串(格式 2006-01-02 15:04:05),可直接字符串比较。
// 注意: xorm SameMapper 产生 PascalCase 列名,列为 CreatedAt 非 created_at。
func cleanupOldLogs() {
	cutoff := time.Now().AddDate(0, 0, -7).Format("2006-01-02 15:04:05")
	if _, err := DB.Where("CreatedAt < ?", cutoff).Delete(new(Log)); err != nil {
		logs.Errorf("清理过期日志失败: %v", err)
	}
}

// noticeWS 通过 WebSocket 推送执行结果通知
// 多个任务可能并发执行(各自 goroutine),同时写同一 ws 连接会触发
// fasthttp/websocket 的 "concurrent write" panic,故每连接加互斥锁序列化写。
func noticeWS(success bool, name string, errMsg string) {
	msg := fmt.Sprintf("[%s] 执行成功", name)
	if !success {
		msg = fmt.Sprintf("[%s] 执行失败: %s", name, errMsg)
	}
	WS.Range(func(_ *fbr.Websocket, client *wsClient) bool {
		client.mu.Lock()
		defer client.mu.Unlock()
		client.conn.WriteJSON(map[string]interface{}{
			"type":    "exec_result",
			"success": success,
			"msg":     msg,
		})
		return true
	})
}
