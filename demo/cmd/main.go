package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	Nats_Urls string
	Nats_User string
	Nats_Pass string
)

func init() {
	// log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetFlags(0)
}

func main() {
	var help bool
	flag.StringVar(&Nats_Urls, "s", "nats://127.0.0.1:4222,nats://127.0.0.1:4223,nats://127.0.0.1:4224", "nats server url")
	flag.StringVar(&Nats_User, "u", "app", "nats username")
	flag.StringVar(&Nats_Pass, "p", "app", "nats password")
	flag.BoolVar(&help, "h", false, "help")
	flag.Usage = func() {
		fmt.Println(`
Usage: flag [-s nats url] [-u nats user] [-p name password]
 
Options:
    `)
		flag.PrintDefaults()
	}
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
		return
	}

	closeCh := make(chan os.Signal, 1)
	nc, cancel := createNc()
	defer cancel()
	go loop(nc)
	signal.Notify(closeCh, syscall.SIGTERM, syscall.SIGINT)
	for sig := range closeCh {
		log.Println("signal:", sig)
		return
	}
}
func Try(f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error:%v,\nStack:%v\n", err, dumpStack(2))
		}
	}()
	f()
}

func loop(nc *nats.Conn) {
	log.Println("输入命令：'AppName CMD ARGES' eg:MONITOR ip")
	scanner := bufio.NewScanner(os.Stdin)
	Try(func() {
		for scanner.Scan() {
			Try(func() {
				result := scanner.Text()
				args := strings.Split(result, " ")
				if len(args) > 1 {
					appName := args[0]
					subj := fmt.Sprintf("CMD.EXC.%v", appName)
					cmdStr := strings.Join(args[1:], " ")
					log.Printf("CMD: %v %v\n", appName, cmdStr)
					rep, err := nc.Request(subj, []byte(cmdStr), 30*time.Second)
					if err != nil {
						log.Printf("respone err:%v\n", err)
					} else {
						log.Printf("respone data:%v\n", string(rep.Data))
					}
				} else {
					log.Printf("invalid cmd:%v\n", result)
				}
			})
		}
	})

}

// Connect to NATS
func createNc() (nc *nats.Conn, deferFunc func()) {
	var err error

	log.Printf("createNc Nats_Urls:%v,Nats_User:%v Nats_Pass:%v\n", Nats_Urls, Nats_User, Nats_Pass)

	nc, err = nats.Connect(Nats_Urls,
		nats.UserInfo(Nats_User, Nats_Pass),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			log.Printf("nats disconnected err:%v,status:%v servers:%v\n", err, conn.Status().String(), conn.Servers())
		}),
		nats.ConnectHandler(func(conn *nats.Conn) {
			// log.Printf("nats connected to %+v\n", conn)
			log.Printf("nats connected to %v addr:%v cluster:%v id:%v status:%v servers:%v\n", conn.ConnectedServerName(), conn.ConnectedAddr(), conn.ConnectedClusterName(), conn.ConnectedServerId(), conn.Status().String(), conn.Servers())
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			log.Printf("nats connection closed,status:%v servers:%v\n", conn.Status().String(), conn.Servers())
		}),
		nats.ErrorHandler(func(conn *nats.Conn, s *nats.Subscription, err error) {
			if s != nil {
				log.Printf("nats connection async err status:%v servers:%v in %q/%q,err:%v\n", conn.Status().String(), conn.Servers(), s.Subject, s.Queue, err)
			} else {
				log.Printf("nats connection async err status:%v servers:%v err:%v\n", conn.Status().String(), conn.Servers(), err)
			}
		}))
	if err != nil {
		log.Fatalln(err)
		return
	}
	deferFunc = func() {
		nc.Drain()
		log.Printf("nats connection is %v\n", nc.Status().String())
	}
	return
}

func dumpStack(callDepth int) string {
	var buff bytes.Buffer
	for i := callDepth + 1; ; i++ {
		/*funcName*/ _, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buff.WriteString(fmt.Sprintf("\n %d:[%s:%d]", i-callDepth, file, line))
	}
	return buff.String()
}
