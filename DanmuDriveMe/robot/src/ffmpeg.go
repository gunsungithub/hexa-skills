package DanmuDriveMe

/*
#cgo LDFLAGS:-L../deps/lib -lavutil -lavformat -lswscale -lswresample -lavcodec -lm
#cgo CFLAGS:-I../deps/include -DBUILD_CGO
#include <stdlib.h>
#include <string.h>
#include <ffmpeg.h>
extern int status[10];
*/
import "C"
import (
	"mind/core/framework/log"
	"mind/core/framework/drivers/media"
	"mind/core/framework/drivers/audio"
	"unsafe"
	"sync"
	"time"
)

type Slice struct {
	Data []byte
	data *c_slice_t
}

type Slice16 struct {
	Data []uint16
	data *c_slice_t
}

type c_slice_t struct {
	p unsafe.Pointer
	n int
}

var frame_index int = 0
var t float64 = 0
var tincr float64 = 2 * 3.141592653 * 110.0 / 44100
var tincr2 float64 = 2 * 3.141592653 * 110.0 / 44100 / 44100

//export fill_image_bytes_GO
func fill_image_bytes_GO(Y, Cb, Cr unsafe.Pointer, width, height int) {
	image := media.SnapshotYCbCr()
	if (width != image.Rect.Max.X || height != image.Rect.Max.Y || image.SubsampleRatio != 2) {
		log.Error.Println(image.Rect, image.SubsampleRatio)
		return
	}
	C.memcpy(Y, unsafe.Pointer(&image.Y[0]),C.size_t(len(image.Y)))
	C.memcpy(Cb, unsafe.Pointer(&image.Cb[0]),C.size_t(len(image.Cb)))
	C.memcpy(Cr, unsafe.Pointer(&image.Cr[0]),C.size_t(len(image.Cr)))
}

//export fill_audio_bytes_GO
func fill_audio_bytes_GO(buf unsafe.Pointer, nb_samples, channels int) {
	readBuf, err := audio.Read()
	log.Debug.Println(len(readBuf), nb_samples*channels)
	if err != nil || len(readBuf) != nb_samples * channels {
		log.Error.Println("Audio read error:", err)
		return
	}
	C.memcpy(buf, unsafe.Pointer(&readBuf[0]),C.size_t(len(readBuf)))
}

func stream_start(url string) {
	cc := C.CString(url)
	defer C.free(unsafe.Pointer(cc))
	if stream_control(true) {
		log.Debug.Println("url:["+url+"]")
		go C.setup(cc)
		// avoid to free cc too early
		time.Sleep(time.Second)
	}
}

var stream_stop bool = true
var mutex sync.Mutex

//export is_stream_stop
func is_stream_stop() C.int {
	if stream_stop {
		log.Debug.Println("is_stream_stop", stream_stop)
		return C.int(1)
	}
	return C.int(0)
}


func stream_control(run bool) bool {
	mutex.Lock()
	defer mutex.Unlock()
	if !stream_stop && run{
		return false
	}
	stream_stop = !run
	if run {
		frame_index = 0
		t = 0
		tincr = 2 * 3.141592653 * 110.0 / 44100
		tincr2 = 2 * 3.141592653 * 110.0 / 44100 / 44100
	}
	log.Debug.Println("stream_control", run)
	return true
}

//export CLog
func CLog(msg *C.char) {
	
	log.Debug.Println(C.GoString(msg))
}
