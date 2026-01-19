package main

import (
	"fmt"
	"net"

	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/logs"
)

var (
	mac = cfg.GetString("mac")
)

func main() {

	fmt.Println("========================================================")
	logs.Info(mac)

	hw, err := net.ParseMAC(mac)
	if err != nil {
		logs.Err(err)
		return
	}

	buf := make([]byte, 102)

	// 前6字节 FF
	for i := 0; i < 6; i++ {
		buf[i] = 0xFF
	}

	// 16次 MAC
	for i := 0; i < 16; i++ {
		copy(buf[6+i*6:], hw)
	}

	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		logs.Err(err)
		return
	}
	defer conn.Close()

	_, err = conn.Write(buf)
	if err != nil {
		logs.Err(err)
		return
	}

	fmt.Println("已发送唤醒命令")

}
