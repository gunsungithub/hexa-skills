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
	C.start(cc)
	C.free(unsafe.Pointer(cc))
}

func streamStop() {
	C.stop()
}

func setText(t string){
	log.Debug.Println(t)
	cc := C.CString(t)
	C.setText(cc)
	C.free(unsafe.Pointer(cc))
}
