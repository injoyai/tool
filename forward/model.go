package forward

import (
	"context"

	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/proxy/core"
	"github.com/injoyai/proxy/forward"
)

type Forward struct {
	ID      int64  `json:"id"`               //主键
	Name    string `json:"name"`             //名称
	Port    int    `json:"port"`             //监听端口
	Address string `json:"address"`          //转发地址
	Enable  bool   `json:"enable"`           //是否启用
	Running bool   `json:"running" xorm:"-"` //是否运行
	Error   string `json:"error" xorm:"-"`   //错误信息
	r       *chans.Rerun
}

func (this *Forward) resp() {
	val, ok := Cache.Get(this.ID)
	if ok {
		this.Running = val.Running
		this.Error = val.Error
	}
}

func (this *Forward) init() {
	this.r = chans.NewRerun(func(ctx context.Context) {
		f := &forward.Forward{
			Listen:  core.NewListenTCP(this.Port),
			Forward: core.NewDialTCP(this.Address),
		}
		this.Running = true
		err := f.Run(ctx)
		this.Running = false
		this.Error = conv.String(err)
	})
	if this.Enable {
		this.r.Enable()
	}
}

func (this *Forward) enable() {
	this.Enable = true
	this.r.Enable()
}

func (this *Forward) disable() {
	this.Enable = false
	this.r.Disable()
}
