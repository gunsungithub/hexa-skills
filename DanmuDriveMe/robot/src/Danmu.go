package DanmuDriveMe

import (
	//"mind/core/framework"
	//"mind/core/framework/skill"
	"errors"
	"io/ioutil"
	"math/rand"
	"mind/core/framework/drivers/hexabody"
	"mind/core/framework/log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	dmServerLabel           = "dm_server"
	dmPortLabel             = "dm_port"
	protocolHeaderSize      = 16
	protocolBodyViewersSize = 4
	sendBufferSize          = 80
	receiveBufferSize       = 256
	magic                   = 16
	protocolVer             = 1
)

var conn *net.TCPConn
var working bool = true

func (d *DanmuDriveMe) connect(_data string) {
	roomid := _data[len("connect:"):]
	dmAddr, err := getDmAddr(roomid)
	if err != nil {
		log.Debug.Println(err)
		return
	}
	tcpaddr, err := net.ResolveTCPAddr("tcp4", dmAddr)
	conn, err = net.DialTCP("tcp", nil, tcpaddr)

	rand.Seed(time.Now().Unix())
	tmpuid := strconv.FormatInt(int64(1e14+2e14*rand.Float64()), 10)

	// send link info
	data, err := generatePacket(0, 7, 1, `{"roomid":`+roomid+`,"uid":`+tmpuid+`}`)
	if err != nil {
		log.Debug.Println(err)
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		log.Debug.Println(err)
		return
	}
	// heart beat
	go heartBeat(30) // 30 seconds

	// receive
	go d.receiveLoop()
}

func heartBeat(delay time.Duration) {
	//waitgroup.Add(1)
	failCount := 0
	var data []byte
	var err error
	for working {
		data, err = generatePacket(0, 2, 1, "")
		if err == nil {
			_, err = conn.Write(data)
			if err == nil {
				log.Debug.Println("heart beat.")
				time.Sleep(delay * time.Second)
				failCount = 0
				continue
			}
		}
		log.Debug.Println("heart beat sent fail times=", failCount)
		if failCount > 5 {
			log.Debug.Println("heart beat sent fail, retry heart beat after 5 seconds.")
			time.Sleep(5 * time.Second)
		}
		failCount++
	}
	log.Debug.Println("heart beat exit.")
	//waitgroup.Done()
}

func (d *DanmuDriveMe) receiveLoop() {
	//waitgroup.Add(1)
	tmpBuffer := make([]byte, 0)

	//声明一个管道用于接收解包的数据
	playerCmdChannel := make(chan []byte, receiveBufferSize)
	go d.parserPlayerCmd(playerCmdChannel)

	buffer := make([]byte, receiveBufferSize)
	for working {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Debug.Println(conn.RemoteAddr().String(), " connection error: ", err)
			return
		}
		log.Debug.Println("n=", n)
		if n > 0 {
			var needMore bool
			tmpBuffer, needMore = Unpack(append(tmpBuffer, buffer[:n]...), playerCmdChannel)
			if needMore {
				continue
			}
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}
	log.Debug.Println("receive exit.")
	//waitgroup.Done()
}

func Unpack(buffer []byte, readerChannel chan []byte) ([]byte, bool) {
	bufferLength := len(buffer)
	//log.Debug.Println("buffer length", bufferLength, buffer)
	if bufferLength > 1000 {
		return make([]byte, 0), true
	}
	if bufferLength < protocolHeaderSize {
		// not enough for a header
		return buffer, true
	} else {
		packetLength, action, _, err := parserHeader(buffer)
		if err != nil {
			log.Debug.Println(err)
			return make([]byte, 0), true
		}
		if bufferLength >= packetLength {
			handleRead(action, buffer[protocolHeaderSize:packetLength], readerChannel)
			return buffer[packetLength:], (bufferLength-packetLength < protocolHeaderSize)
		} else {
			log.Debug.Println("need more bytes to make a packet")
			return buffer, true
		}
	}
}

func getDmAddr(roomId string) (string, error) {
	resp, err := http.Get("http://live.bilibili.com/api/player?id=cid:" + roomId)
	if err != nil {
		log.Debug.Println(err)
		return "", errors.New("request bilibili api/player fail with id=" + roomId)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debug.Println(err)
		return "", errors.New("read bilibili api/player fail with id=" + roomId)
	}
	dmAddr := ""
	reg := regexp.MustCompile(dmServerLabel + `(.*)` + dmServerLabel)
	result := reg.FindString(string(body)) // find danmu server
	if len(result) > len(dmServerLabel+">"+"</"+dmServerLabel) {
		dmAddr = result[len(dmServerLabel+">"):len(result)-len("</"+dmServerLabel)] + ":"
	} else {
		return "", errors.New("search danmu server fail with id=" + roomId)
	}
	reg = regexp.MustCompile(dmPortLabel + `(.*)` + dmPortLabel)
	result = reg.FindString(string(body)) // find danmu port
	if len(result) > len(dmPortLabel+">"+"</ "+dmPortLabel) {
		dmAddr = dmAddr + result[len(dmPortLabel+">"):len(result)-len("</"+dmPortLabel)]
	} else {
		return "", errors.New("search danmu port fail with id=" + roomId)
	}
	return dmAddr, nil
}

func generatePacket(packetlength int, action int, param int, body string) ([]byte, error) {
	if packetlength == 0 || packetlength < protocolHeaderSize {
		packetlength = len(body) + protocolHeaderSize
	}
	var buffer [sendBufferSize]byte
	buffer[0] = byte((packetlength >> 24) & 0xFF)
	buffer[1] = byte((packetlength >> 16) & 0xFF)
	buffer[2] = byte((packetlength >> 8) & 0xFF)
	buffer[3] = byte(packetlength & 0xFF)

	buffer[4] = byte((magic >> 8) & 0xFF) // magic
	buffer[5] = byte(magic & 0xFF)

	buffer[6] = byte((protocolVer >> 8) & 0xFF) // ver
	buffer[7] = byte(protocolVer & 0xFF)

	buffer[8] = byte((action >> 24) & 0xFF) // action
	buffer[9] = byte((action >> 16) & 0xFF)
	buffer[10] = byte((action >> 8) & 0xFF)
	buffer[11] = byte(action & 0xFF)

	buffer[12] = byte((param >> 24) & 0xFF) // param
	buffer[13] = byte((param >> 16) & 0xFF)
	buffer[14] = byte((param >> 8) & 0xFF)
	buffer[15] = byte(param & 0xFF)

	if packetlength > protocolHeaderSize && packetlength < sendBufferSize {
		playload := []byte(body)
		length := packetlength - protocolHeaderSize
		for i := 0; i < length; i++ {
			buffer[i+protocolHeaderSize] = playload[i]
		}
	} else if packetlength >= sendBufferSize {
		log.Debug.Println("body"+body+", please limit to ", sendBufferSize-protocolHeaderSize)
		return nil, errors.New("body is too large")
	}
	return buffer[:packetlength], nil
}

func getValue(data []byte) int {
	value := 0
	if len(data) > 0 {
		value = int(data[0])
	} else {
		return 0
	}
	for i := 1; i < len(data); i++ {
		value <<= 8
		value += int(data[i])
	}
	return value
}

func parserHeader(header []byte) (int, int, int, error) {
	length := getValue(header[:4])
	if length < 16 {
		log.Debug.Println("packet size ", length, "<", protocolHeaderSize)
		return 0, 0, 0, errors.New("parser header fail")
	}
	m := getValue(header[4:6])
	v := getValue(header[6:8])
	if m != magic || (v != protocolVer && v != 0) {
		log.Debug.Println(m, "!=", magic, " || ", v, "!=", protocolVer)
		return 0, 0, 0, errors.New("parser header fail")
	}
	action := getValue(header[8:12])
	if action == 3 && length != protocolHeaderSize+protocolBodyViewersSize {
		log.Debug.Println("packet size", length, "should be", protocolHeaderSize, "with action ", action)
		return 0, 0, 0, errors.New("parser header fail")
	}
	if action == 8 && length != protocolHeaderSize {
		log.Debug.Println("packet size", length, "should be", protocolHeaderSize, "with action ", action)
		return 0, 0, 0, errors.New("parser header fail")
	}
	param := getValue(header[12:16])
	return length, action, param, nil
}

func parserViewers(body []byte) (int, error) {
	if len(body) < protocolBodyViewersSize {
		log.Debug.Println("body size ", len(body), "<", protocolBodyViewersSize)
		return 0, errors.New("parser viewers fail")
	}
	viewers := getValue(body[:4])
	return viewers, nil
}

func (d *DanmuDriveMe) parserPlayerCmd(readerChannel chan []byte) {
	regCmd := regexp.MustCompile(`"cmd":"[A-Z]*_[A-Z]*"`)
	regDanmu := regexp.MustCompile(`\],".*",\[`)
	for {
		select {
		case data := <-readerChannel:

			s := string(data)
			cmd := regCmd.FindString(s)
			length := len(cmd)
			if length < len(`"cmd":""`) {
				log.Debug.Println("unknow cmd", cmd, s)
				continue
			}
			switch cmd[7 : length-1] {
			case "DANMU_MSG":
				danmu := regDanmu.FindString(s)
				length = len(danmu)
				if length > len(`],"",[`) {
					danmu = strings.TrimSpace(danmu[3 : length-3])
					log.Debug.Println(danmu)
					if danmu == "a" {
						hexabody.MoveHead(hexabody.Direction()+45, 10)
					} else if danmu == "d" {
						hexabody.MoveHead(hexabody.Direction()-45, 10)
					} else if danmu == "w" {
						hexabody.Walk(hexabody.Direction(), 50)
					} else if danmu == "s" {
						du := hexabody.Direction() + 180
						if du > 360 {
							du -= 360
						}
						hexabody.Walk(du, 50)
					} else if danmu == "l" {
						du := hexabody.Direction() + 90
						if du > 360 {
							du -= 360
						}
						hexabody.Walk(du, 50)
					} else if danmu == "r" {
						du := hexabody.Direction() + 270
						if du > 360 {
							du -= 360
						}
						hexabody.Walk(hexabody.Direction(), 50)
					}
				}
			case "SYS_GIFT":
				log.Debug.Println("SYS_GIFT")
			case "SYS_MSG":
				log.Debug.Println("SYS_MSG")
			default:
				log.Debug.Println("unknow cmd", s, cmd)
			}
		}
	}
}

func handleRead(action int, body []byte, readerChannel chan []byte) error {
	var err error = nil
	switch action {
	case 3: //number of viewers
		viewers, err := parserViewers(body[:protocolBodyViewersSize])
		if err == nil {
			log.Debug.Println("Number of viewers = ", viewers)
		}
	case 5:
		readerChannel <- body
	case 8:
		log.Debug.Println("Entry room succeed!")
	default:
		log.Debug.Println("receive unknow action:", action)
		err = errors.New("parser body fail")
	}
	return err
}
