package timer

import (
	"fmt"

	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/task"
)

type Timer struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Cron    string `json:"cron"`
	Content string `json:"content"`
	Enable  bool   `json:"enable"`
	Next    string `json:"next" xorm:"-"`
	ExecErr string `json:"execErr" xorm:"-"`
}

func (this *Timer) Resp(t *task.Task[string]) {
	if t != nil {
		this.Next = t.Next.Format("2006-01-02 15:04:05")
	}
}

func (this *Timer) String() string {
	return fmt.Sprintf("[%s][%s][%02d:%s] %s", conv.Select(this.Enable, "启用", "禁用"), this.Cron, this.ID, this.Name, this.Content)
}

func (this *Timer) ExecText() string {
	return fmt.Sprintf("[执行][%s][%02d:%s] %s", this.Cron, this.ID, this.Name, this.Content)
}

// Setting 通用 KV 配置(错误处理脚本等全局设置)
type Setting struct {
	Key   string `json:"key"   xorm:"pk"`
	Value string `json:"value"`
}
