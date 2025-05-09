package utils

import (
	"context"
	"sync"
)

// 全局变量
var (
	Gctx      context.Context
	Gwg       *sync.WaitGroup
	Gshutdown context.CancelFunc
	AppName   string
	Nats      *NatsObj
)
