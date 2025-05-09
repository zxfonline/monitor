package trace

import (
	"github.com/zxfonline/monitor/trace/expvar"
	"github.com/zxfonline/monitor/trace/golangtrace"
)

var Enable = true

type ProxyTrace struct {
	tr golangtrace.Trace
}

func Start(family, title string, expvar bool) *ProxyTrace {
	if Enable {
		pt := &ProxyTrace{tr: golangtrace.New(family, title, expvar)}
		return pt
	}
	return nil
}

func Finish(pt *ProxyTrace) {
	if pt != nil {
		if pt.tr != nil {
			pt.tr.Finish()
		}
	}
}

func Finish2Expvar(pt *ProxyTrace, traceDefer func(*expvar.Map, int64)) {
	if pt != nil {
		if pt.tr != nil {
			family := pt.tr.GetFamily()
			elapsed := pt.tr.Finish()
			if traceDefer != nil {
				req := expvar.Get(family)
				if req == nil {
					func() {
						defer func() {
							if err := recover(); err != nil {
								req = expvar.Get(family)
							}
						}()
						req = expvar.NewMap(family)
					}()
				}
				traceDefer(req.(*expvar.Map), elapsed.Nanoseconds())
			}
		}
	}
}

func Printf(pt *ProxyTrace, format string, a ...any) {
	if pt != nil {
		if pt.tr != nil {
			pt.tr.LazyPrintf(format, a...)
		}
	}
}

func Errorf(pt *ProxyTrace, format string, a ...any) {
	if pt != nil {
		if pt.tr != nil {
			pt.tr.LazyPrintf(format, a...)
			pt.tr.SetError()
		}
	}
}

// RegisterChanMonitor 注册管道监控
func RegisterChanMonitor(name string, chanPtr any) bool {
	if !isChan(chanPtr) {
		return false
	}
	monitorLock.Lock()
	defer monitorLock.Unlock()
	if monitorChanMap[name] != nil {
		return false
	}
	monitorChanMap[name] = chanPtr
	return true
}

// DeleteChanMonitor 删除管道监控
func DeleteChanMonitor(name string) {
	monitorLock.Lock()
	defer monitorLock.Unlock()
	delete(monitorChanMap, name)
}
