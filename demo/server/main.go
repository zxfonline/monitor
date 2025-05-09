package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zxfonline/monitor/trace"
	"github.com/zxfonline/monitor/trace/expvar"
	"github.com/zxfonline/monitor/trace/golangtrace"
	"github.com/zxfonline/monitor/utils"
)

const (
	cfgPath = "../app.conf"
)

func init() {
	utils.Gwg = &sync.WaitGroup{}
	utils.Gctx, utils.Gshutdown = context.WithCancel(context.Background())
	utils.InitIniFile(cfgPath)

	utils.AppName = utils.IniReadString("GOLBAL_CONF", "app_name")
	{
		nats_url := utils.IniReadString("GOLBAL_CONF", "nats_base_url")
		if len(nats_url) > 0 {
			nats_user := utils.IniReadString("GOLBAL_CONF", "nats_base_user")
			nats_pwd := utils.IniReadString("GOLBAL_CONF", "nats_base_pwd")
			utils.Nats = utils.CreateNatsObj(nats_url, nats_user, nats_pwd, false)
			if utils.Nats == nil {
				log.Println("BASE NatsObj创建失败")
			}
		}
	}
}

func main() {
	utils.SubCMDHandler()
	msgChan := make(chan int, 10000)
	{
		trace.RegisterChanMonitor("chanMainProc", msgChan)
		msgName := "Func:main"
		proxyTrace := trace.Start("Actor1", msgName, true)
		defer trace.Finish2Expvar(proxyTrace, func(req *expvar.Map, time int64) {
			req.AddMessage(msgName, 1, time)
		})
		trace.Printf(proxyTrace, "I'm trace")
		time.Sleep(1 * time.Second)
		trace.Errorf(proxyTrace, "Oh no, there was a mistake")

		event := golangtrace.NewEventLog("Event1", msgName)
		event.Printf("I'm event")
		time.Sleep(1 * time.Second)
		event.Errorf("mistake")
		time.Sleep(1 * time.Second)
		defer event.Finish()
	}
	pprofHandler := http.NewServeMux()
	trace.Init(pprofHandler, "test")
	strIp := utils.IniReadString(utils.AppName, "localIP")
	nPort := utils.IniReadInt64(utils.AppName, "httpport")
	srv := &http.Server{Addr: fmt.Sprintf("%v:%v", strIp, nPort), Handler: pprofHandler}
	go func() {
		ln, err := net.Listen("tcp", srv.Addr)
		if err == nil {
			log.Printf("http://127.0.0.1:%v/debug/pprof/\n", nPort)
			srv.ListenAndServe()
			err = srv.Serve(ln)
			if err == nil {
				log.Printf("PPROF启动成功, http server listen:%v\n", nPort)
			} else if err != http.ErrServerClosed {
				log.Printf("PPROF启动ERROR, http server listen:%v 错误:%v\n", nPort, err.Error())
			}
		}

	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		select {
		case sys_signal := <-ch:
			log.Printf("Receive OS Signal: %v\n", sys_signal.String())
			switch sys_signal {
			case syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				// case syscall.SIGTERM:
				srv.Shutdown(context.Background())
				return
			}
		}
	}
}
