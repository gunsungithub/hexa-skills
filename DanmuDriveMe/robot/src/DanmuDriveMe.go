package DanmuDriveMe

import (
	"mind/core/framework"
	"mind/core/framework/drivers/hexabody"
	"mind/core/framework/drivers/media"
	"mind/core/framework/drivers/audio"
	"mind/core/framework/log"
	"mind/core/framework/skill"
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
	stream_control(false)
}

func (d *DanmuDriveMe) OnConnect() {
	log.Debug.Println("OnConnect")
	if !media.Available() {
		log.Error.Println("Media driver not available")
		return
	}
	if err := media.Start(); err != nil {
		log.Error.Println("Media driver could not start")
	}
	if !audio.Available() {
		log.Error.Println("Audio driver not available")
		return
	}
	if err := audio.Start(); err != nil {
		log.Error.Println("Audio driver could not start")
		return
	}
	audio.Init(1, 44100, audio.FormatS16LE)
}

func (d *DanmuDriveMe) OnDisconnect() {
	log.Debug.Println("OnDisconnect")
	if conn != nil {
		conn.Close()
	}
	audio.Close()
	os.Exit(0)
}

func (d *DanmuDriveMe) OnRecvJSON(data []byte) {
	log.Debug.Println("OnRecvJSON", string(data))
}

func (d *DanmuDriveMe) OnRecvString(data string) {
	log.Debug.Println("OnRecvString", data)
	framework.SendString(data)
	switch data {
	default:
		if len(data) > len("connect") && data[:len("connect")] == "connect" {
			d.connect(data)
		}
		if len(data) > len("stream_send") && data[:len("stream_send")] == "stream_send" {
			stream_start(data[len("stream_send")+1:])
		}
	case "finish":
		working = false
	case "stream_stop":
		stream_control(false)
	}
}
