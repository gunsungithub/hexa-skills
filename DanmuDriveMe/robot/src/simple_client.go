package DanmuDriveMe

/*
#cgo CFLAGS:-I../deps/include
#include <stdlib.h>
#include <string.h>
#include <simple_client.h>
*/
import "C"
import (
	"mind/core/framework/log"
	"time"
	"unsafe"
)

func stream_start(url string) {
	cc := C.CString(url)
	defer C.free(unsafe.Pointer(cc))
	if stream_control(true) {
		log.Debug.Println("url:[" + url + "]")
		C.setup(cc)
	}
}

var stream_stop bool = true

func stream_control(run bool) bool {
	if !stream_stop && run {
		return false
	}
	stream_stop = !run
	if run {
		go heartb(5)
	}
	log.Debug.Println("stream_control", run)
	return true
}

//export CLog
func CLog(msg *C.char) {
	log.Debug.Println(C.GoString(msg))
}

// heart beat
func heartb(delay time.Duration) {
	run := C.CString("1")
	stop := C.CString("0")
	defer C.free(unsafe.Pointer(run))
	defer C.free(unsafe.Pointer(stop))
	for !stream_stop {
		C.setup(run)
		log.Debug.Println("alive")
		time.Sleep(delay * time.Second)
	}
	log.Debug.Println("finish")
	C.setup(stop)
}
