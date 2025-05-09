package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"

	"sort"
	"strings"
)

type (
	callBack struct {
		f    func([]string)
		main bool
	}
	cmd_callback struct {
		callback map[string]*callBack
	}
)

var Cmd = cmd_callback{callback: make(map[string]*callBack, 0)}

func init() {
	go loop()
	Cmd.Regist("h", Cmd.help, true)
}

func loop() {
	scanner := bufio.NewScanner(os.Stdin)
	Try(func() {
		for scanner.Scan() {
			Try(func() {
				result := scanner.Text()
				args := strings.Split(result, " ")
				cmdName := args[0]
				node := Cmd.callback[cmdName]
				if node == nil {
					log.Printf("无效的命令:%v\n", cmdName)
				} else {
					f := node.f
					main := node.main
					if f != nil {
						if main {
							//send to actor
							f(args[1:])
						} else {
							f(args[1:])
						}
					} else {
						log.Printf("无效的命令:%v\n", cmdName)
					}
				}
			})
		}
		log.Println("ERROR loop finish")
	})
}

func (sf *cmd_callback) help(sp []string) {
	s := []string{}
	for key := range sf.callback {
		s = append(s, key)
	}

	sort.Strings(s)
	p := strings.Builder{}
	p.WriteString(fmt.Sprintf("%v:所有命令:\n", AppName))
	for _, str := range s {
		p.WriteString(fmt.Sprintf("    %v\n", str))
	}

	p.WriteString(fmt.Sprintf("================== %v\n", AppName))
	fmt.Printf("%v\n", p.String())
}

func (sf *cmd_callback) Regist(cmd string, f func([]string), deal_main bool) {
	if _, exist := sf.callback[cmd]; exist {
		log.Printf("重复注册命令:%v\n", cmd)
		return
	}
	c := &callBack{}
	c.f = f
	c.main = deal_main
	sf.callback[cmd] = c
}

func (sf *cmd_callback) ExistsByCmd(cmd string) bool {
	_, exists := sf.callback[cmd]
	return exists
}

func (sf *cmd_callback) RunCmd(cmd string, args []string) {
	if callback, exists := sf.callback[cmd]; exists {
		callback.f(args)
		return
	}
	log.Printf("命令不存在:%v\n", cmd)
}
func ExcCmd(result string) string {
	args := strings.Split(result, " ")
	cmdName := args[0]
	node := Cmd.callback[cmdName]
	if node == nil {
		log.Printf("无效的命令:%v\n", cmdName)
		return fmt.Sprintf("无效的命令:%v", cmdName)
	} else {
		f := node.f
		main := node.main
		if f != nil {
			if main {
				f(args[1:])
				return "async execution finish"
			} else {
				f(args[1:])
				return "sync execute finish"
			}
		} else {
			log.Printf("无效的命令:%v\n", cmdName)
			return fmt.Sprintf("无效的命令:%v", cmdName)
		}
	}
}
func SubCMDHandler() {
	if Nats == nil {
		return
	}
	strSubj := fmt.Sprintf("CMD.EXC.%v", AppName)
	_, err := Nats.Subscribe(strSubj, func(msg *nats.Msg) {
		result := "无效的命令:"
		if len(msg.Data) != 0 {
			result = func() string {
				defer func() {
					if err := recover(); err != nil {
						log.Printf("Error:%v\n", err)
					}
				}()
				cmdStr := string(msg.Data)
				return ExcCmd(cmdStr)
			}()
		}
		msg.Respond([]byte(result))
	})
	if err != nil {
		log.Printf("[%v] sub err:%v\n", strSubj, err)
	} else {
		log.Printf("[%v] sub ok\n", strSubj)
	}
}
