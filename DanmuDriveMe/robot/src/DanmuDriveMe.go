package DanmuDriveMe

import (
	"mind/core/framework"
	"mind/core/framework/skill"
	"mind/core/framework/log"
	"mind/core/framework/drivers/hexabody"
	"os"
)

type DanmuDriveMe struct {
	skill.Base
	stop chan bool
}

func NewSkill() skill.Interface {
	return &DanmuDriveMe{
		stop: make(chan bool),
	}
}

func (d *DanmuDriveMe) OnStart() {
	log.Debug.Println("OnStart")
	err := hexabody.Start()
	if err != nil {
		log.Error.Println("hexabody can't start:", err)
	}
}

func (d *DanmuDriveMe) OnClose() {
	log.Debug.Println("OnClose")
	hexabody.Close()
}

func (d *DanmuDriveMe) OnConnect() {
	log.Debug.Println("OnConnect")
	
}

func (d *DanmuDriveMe) OnDisconnect() {
	log.Debug.Println("OnDisconnect")
	if conn != nil {
		conn.Close()
	}
	os.Exit(0)
}

func (d *DanmuDriveMe) OnRecvJSON(data []byte) {
	log.Debug.Println("OnRecvJSON", string(data))
}

func (d *DanmuDriveMe) OnRecvString(data string) {
	log.Debug.Println("OnRecvString", data)
	framework.SendString(data)
	switch (data) {
		default:
		if len(data) > len("connect") && data[:len("connect")] == "connect" {
			d.connect(data)
		}
		case "finish":
			working = false
	}
}