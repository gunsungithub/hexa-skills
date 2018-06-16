package DanmuDriveMe

/*
#cgo CFLAGS: -I ../deps/inlcude
#include <stdlib.h>
#include "gst_go.h"
*/
import "C"
import (
	"mind/core/framework/log"
	"unsafe"
)

func streamStart(url string) {
	log.Debug.Println(url)
	cc := C.CString(url)
	C.open_client()
	C.start_client(cc)
	C.free(unsafe.Pointer(cc))
}

func streamStop() {
	C.stop_client()
	C.close_client()
}

func setText(t string) {
	log.Debug.Println(t)
	cc := C.CString(t)
	//C.setText(cc)
	C.free(unsafe.Pointer(cc))
}

//export CLog
func CLog(msg *C.char) {
	log.Debug.Println(C.GoString(msg))
}
