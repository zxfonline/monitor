package expvar

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
)

// Message 消息
type Message struct {
	Num     int64
	MinTime int64
	MaxTime int64
}

func (v *Message) String() string {
	return fmt.Sprintf(`{"Num":%v,"MaxTime":%v,"MinTime":%v}`, atomic.LoadInt64(&v.Num), atomic.LoadInt64(&v.MaxTime), atomic.LoadInt64(&v.MinTime))
}

func (v *Message) Add(delta int64) {
	atomic.AddInt64(&v.Num, delta)
}

func (v *Message) Set(value int64) {
	atomic.StoreInt64(&v.Num, value)
}

func (v *Message) SetTime(time int64) {
	minTime := atomic.LoadInt64(&v.MinTime)
	if minTime > time {
		atomic.CompareAndSwapInt64(&v.MinTime, minTime, time)
	}

	maxTime := atomic.LoadInt64(&v.MaxTime)
	if maxTime < time {
		atomic.CompareAndSwapInt64(&v.MaxTime, maxTime, time)
	}
}

func (v *Map) AddMessage(key string, delta, time int64) {
	i, ok := v.m.Load(key)
	if !ok {
		var dup bool
		i, dup = v.m.LoadOrStore(key, &Message{
			MinTime: 100 * 1e9, //设置100s
		})
		if !dup {
			v.addKey(key)
		}
	}

	if iv, ok := i.(*Message); ok {
		iv.Add(delta)
		iv.SetTime(time)
	}
}

func ExpvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(vars.appendJSONMayExpand(nil, true))
}
func Cmdline() any {
	return os.Args
}

func Memstats() any {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return *stats
}
