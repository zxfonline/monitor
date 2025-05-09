package trace

import (
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/zxfonline/monitor/trace/golangtrace"
)

var (
	startTime      = time.Now().UTC()
	monitorChanMap = make(map[string]any, 16)
	monitorLock    sync.RWMutex
)

func goroutines() any {
	return runtime.NumGoroutine()
}

func uptime() any {
	uptime := time.Since(startTime)
	return int64(uptime)
}

func traceTotal() any {
	return golangtrace.GetAllExpvarFamily(2)
}
func traceHour() any {
	return golangtrace.GetAllExpvarFamily(1)
}
func traceMinute() any {
	return golangtrace.GetAllExpvarFamily(0)
}

func isChan(a any) bool {
	if a == nil {
		return false
	}

	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Chan {
		return false
	}
	if v.IsNil() {
		return false
	}
	return true
}

func chanInfo(a any) (bool, int, int) {
	if a == nil {
		return false, 0, 0
	}

	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Chan {
		return false, 0, 0
	}
	if v.IsNil() {
		return false, 0, 0
	}
	return true, v.Cap(), v.Len()
}

type ChanInfo struct {
	Cap  int
	Len  int
	Rate float64
}

func chanStats() any {
	monitorLock.RLock()
	defer monitorLock.RUnlock()
	mp := make(map[string]ChanInfo)
	for k, v := range monitorChanMap {
		_, m, i := chanInfo(v)
		mp[k] = ChanInfo{Cap: m, Len: i, Rate: float64(int64((float64(i) / float64(m) * 10000.0))) / 10000.0}
	}
	return mp
}
