package main

import (
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/tool/upgrade"
	"os/exec"
	"path/filepath"
)

func main() {
	name := "in.exe"
	filename := filepath.Join("./", name)
	up := upgrade.Upgrade{
		Filename: filename,
		Retry:    3,
		Backups:  filename + ".bak",
		Stop: func() error {
			return shell.Stop(name)
		},
		Start: func() error {
			return exec.Command("cmd", "/c", "start "+filename).Start()
		},
	}
	up.Upgrade(&upgrade.Info{
		Url:     "http://192.168.1.105:9000/in-store/" + name,
		Version: "v1.0.0",
		MD5:     "",
	})
}
