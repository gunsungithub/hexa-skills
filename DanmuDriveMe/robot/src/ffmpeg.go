package DanmuDriveMe

/*
#cgo LDFLAGS:-L../deps/lib -lavutil -lavformat -lswscale -lswresample -lavcodec -lm
#cgo CFLAGS:-I../deps/include -DBUILD_CGO
#include <stdlib.h>
#include <ffmpeg.h>
extern int status[10];
*/
import "C"
import (
	"mind/core/framework/log"
	"math"
	"reflect"
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
	var x, y int
	var i int
	i = frame_index
	frame_index++
	dataY := &c_slice_t{Y, width * height}
	sY := &Slice{data: dataY}
	hY := (*reflect.SliceHeader)((unsafe.Pointer(&sY.Data)))
	hY.Cap = dataY.n
	hY.Len = dataY.n
	hY.Data = uintptr(Y)
	/* Y */
	for y = 0; y < height; y++ {
		for x = 0; x < width; x++ {
			sY.Data[y*width+x] = (byte)(x + y + i*3)
		}
	}
	dataCb := &c_slice_t{Cb, width * height / 4}
	sCb := &Slice{data: dataCb}
	hCb := (*reflect.SliceHeader)((unsafe.Pointer(&sCb.Data)))
	hCb.Cap = dataCb.n
	hCb.Len = dataCb.n
	hCb.Data = uintptr(Cb)

	dataCr := &c_slice_t{Cr, width * height / 4}
	sCr := &Slice{data: dataCb}
	hCr := (*reflect.SliceHeader)((unsafe.Pointer(&sCr.Data)))
	hCr.Cap = dataCr.n
	hCr.Len = dataCr.n
	hCr.Data = uintptr(Cr)
	/* Cb and Cr */
	for y = 0; y < height/2; y++ {
		for x = 0; x < width/2; x++ {
			sCb.Data[y*(width>>1)+x] = (byte)(128 + y + i*2)
			sCr.Data[y*(width>>1)+x] = (byte)(64 + x + i*5)
		}
	}
}

//export fill_audio_bytes_GO
func fill_audio_bytes_GO(buf unsafe.Pointer, nb_samples, channels int) {
	var j, i int
	var v uint16
	data := &c_slice_t{buf, nb_samples * channels * 2}
	s := &Slice16{data: data}
	h := (*reflect.SliceHeader)((unsafe.Pointer(&s.Data)))
	h.Cap = data.n
	h.Len = data.n
	h.Data = uintptr(buf)

	for j = 0; j < nb_samples; j++ {
		v = (uint16)(math.Sin(t) * 10000)
		for i = 0; i < channels; i++ {
			s.Data[j*channels+i] = v
		}
		t += tincr
		tincr += tincr2
	}
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
