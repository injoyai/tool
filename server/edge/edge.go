package edge

import (
	"errors"
	"fmt"
	"github.com/go-toast/toast"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg/v2"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/shell"
)

func Do(Type string, data any) error {
	switch Type {
	case "edge.notice.upgrade":
		m := conv.NewMap(data)

		//显示通知和是否升级按钮按钮
		upgradeEdge := fmt.Sprintf("http://localhost:%d", cfg.GetInt("http.port")) + "?cmd=i%20server%20edge%20upgrade"
		notification := toast.Notification{
			AppID:   "Microsoft.Windows.Shell.RunDialog",
			Title:   fmt.Sprintf("发现新版本(%s),是否马上升级?", m.GetString("version")),
			Message: "版本详情: " + m.GetString("versionDetails"),
			Actions: []toast.Action{
				{"protocol", "马上升级", upgradeEdge},
				{"protocol", "稍后再说", ""},
			},
		}
		if err := notification.Push(); err != nil {
			return err
		}

		//播放语音
		notice.DefaultVoice.Speak(fmt.Sprintf("主人. 发现网关新版本(%s). 是否马上升级?", m.GetString("version")))

	case "edge.upgrade":
		return shell.Start("i server edge upgrade")

	case "edge.open", "edge.run", "edge.start":
		return shell.Start("i server edge")

	case "edge.close", "edge.stop", "edge.shutdown":
		return shell.Start("i server edge stop")
	}
	return errors.New("未知命令: " + Type)
}
