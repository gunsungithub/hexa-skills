package DanmuDriveMe

/*
#cgo LDFLAGS: -L../deps/lib -lrtmp
#include <stdlib.h>
#include "../deps/include/rtmp_sample_api.c"
#include <sys/times.h>
#include <unistd.h>
*/
import "C"
import (
	"bufio"
	"mind/core/framework/log"
	"os"
	"time"
	"unsafe"
)

var rtmp_working bool = false
var clk_tck C.long = 0

func GetTime() uint {
	var t C.struct_tms
	if clk_tck == 0 {
		clk_tck = C.sysconf(C._SC_CLK_TCK)
		log.Debug.Println("Hello", clk_tck)
	}
	return uint(C.ulong(C.times(&t)) * 1000 / C.ulong(clk_tck))
}
func setup(url string) {
	if rtmp_working {
		return
	}
	rtmp_working = true

	f, err := os.OpenFile("assets/test.flv", os.O_RDONLY, 0666)
	if err != nil {
		log.Debug.Println("OpenFile fail!")
		return
	}
	defer f.Close()
	C.rtmp_sample_init()
	defer C.rtmp_sample_final()
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	C.rtmp_sample_connect(curl)
	defer C.rtmp_sample_disconnect()
	start_time := GetTime()
	var pre_frame_time uint = 0
	var lasttime uint = 0
	bNextIsKey := false
	i := 0
	var datalength uint = 0
	var timestamp uint = 0
	reader := bufio.NewReaderSize(f, 1024*1024)
	reader.Discard(13)
	for rtmp_working {

		//log.Debug.Println("i:", i, bNextIsKey, GetTime(), start_time, pre_frame_time)
		if GetTime()-start_time < pre_frame_time && bNextIsKey {
			if pre_frame_time > lasttime {
				log.Debug.Println("Time Stamp:", pre_frame_time, "ms", i)
				lasttime = pre_frame_time
			}
			//i++
			time.Sleep(time.Second)
			continue
		}
		i++
		//log.Debug.Println("i:", i)

		buf, _ := reader.Peek(8)
		datalength = uint(buf[1]) << 16
		//log.Debug.Println("tmp:", buf[1])
		//tmp, _ = reader.ReadByte()
		datalength += uint(buf[2]) << 8
		//log.Debug.Println("tmp:", buf[2])
		//tmp, _ = reader.ReadByte()
		datalength += uint(buf[3])
		//log.Debug.Println("tmp:", buf[3])
		//log.Debug.Println("datalength:", datalength)
		//tmp, _ = reader.ReadByte()
		timestamp = 0
		timestamp = uint(buf[4]) << 16
		//tmp, _ = reader.ReadByte()
		timestamp += uint(buf[5]) << 8
		//tmp, _ = reader.ReadByte()
		timestamp += uint(buf[6])
		//tmp, _ = reader.ReadByte()
		//timestamp += uint(buf[7])
		//log.Debug.Println("timestamp:", timestamp, buf[4:8])
		buf, err := reader.Peek(int(datalength) + 15)
		if err != nil {
			log.Debug.Println(err, datalength)
			reader.Discard(int(datalength) + 15)
			continue
		}
		//log.Debug.Println("buf:", buf[:5])
		reader.Discard(int(datalength) + 15)
		cc := C.CString(string(buf))
		pre_frame_time = timestamp
		C.rtmp_sample_add_data(cc, C.int(datalength)+15)
		C.free(unsafe.Pointer(cc))
		buf, _ = reader.Peek(12)
		if buf[0] == 0x09 {
			//buf, _ = reader.Peek(1)
			if buf[11] == 0x17 {
				bNextIsKey = true
			} else {
				bNextIsKey = false
			}
		}
	}
}

func rtmp_send (url string) {
	go setup(url)
}
