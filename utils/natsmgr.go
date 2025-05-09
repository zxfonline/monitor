package utils

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type (
	NatsObj struct {
		nc  *nats.Conn
		jet nats.JetStreamContext
	}
)

func CreateStreamConfig(strStreamName string, sSubjects []string, nReplicas int) *nats.StreamConfig {
	return &nats.StreamConfig{
		Name:      strStreamName,
		Subjects:  sSubjects,
		Retention: nats.WorkQueuePolicy,
		Discard:   nats.DiscardOld,
		Storage:   nats.FileStorage,
		Replicas:  nReplicas,
	}
}

func CreateNatsObj(strNatsURL, strNatsUser, strNatsPwd string, isInitJetStream bool) *NatsObj {
	tNatsObj := initNats(strNatsURL, strNatsUser, strNatsPwd, "", isInitJetStream)
	return tNatsObj
}

// 第一参数: nats服务器配置
// 单机模式: nats://localhost:4443
// 集群模式: nats://localhost:1222, nats://localhost:1223, nats://localhost:1224
func initNats(strNatsServer, strUser, strPwd, strToken string, isInitJetStream bool) *NatsObj {
	sOption := make([]nats.Option, 0, 2)
	if len(strUser) > 0 && len(strPwd) > 0 {
		sOption = append(sOption, nats.UserInfo(strUser, strPwd))
	}
	if len(strToken) > 0 {
		sOption = append(sOption, nats.Token(strToken))
	}
	sOption = append(sOption, nats.PingInterval(10*time.Second))
	sOption = append(sOption, nats.MaxPingsOutstanding(3))
	// sOption = append(sOption, nats.Name(core.AppName))
	sOption = append(sOption, nats.Timeout(10*time.Second))
	sOption = append(sOption, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		log.Printf("ERROR client disconnected err:%v,status:%v servers:%v\n", err, conn.Status().String(), conn.Servers())
	}))
	sOption = append(sOption, nats.ReconnectHandler(func(conn *nats.Conn) {
		log.Printf("ERROR client reconnected to %v addr:%v cluster:%v id:%v status:%v servers:%v\n", conn.ConnectedServerName(), conn.ConnectedAddr(), conn.ConnectedClusterName(), conn.ConnectedServerId(), conn.Status().String(), conn.Servers())
	}))
	sOption = append(sOption, nats.ClosedHandler(func(conn *nats.Conn) {
		log.Printf("ERROR client connection closed,status:%v servers:%v\n", conn.Status().String(), conn.Servers())
	}))
	sOption = append(sOption, nats.ConnectHandler(func(conn *nats.Conn) {
		log.Printf("client connected to %v addr:%v cluster:%v id:%v status:%v servers:%v\n", conn.ConnectedServerName(), conn.ConnectedAddr(), conn.ConnectedClusterName(), conn.ConnectedServerId(), conn.Status().String(), conn.Servers())
	}))

	// sOption = append(sOption, nats.DisconnectHandler(func(conn *nats.Conn) {
	// 	log.Printf("ERROR client disconnected from %v addr:%v cluster:%v id:%v status:%v servers:%v\n", conn.ConnectedServerName(), conn.ConnectedAddr(), conn.ConnectedClusterName(), conn.ConnectedServerId(), conn.Status().String(), conn.Servers())
	// }))
	sOption = append(sOption, nats.DiscoveredServersHandler(func(conn *nats.Conn) {
		log.Printf("client discover servers %v addr:%v cluster:%v id:%v status:%v servers:%v discovered:%v\n", conn.ConnectedServerName(), conn.ConnectedAddr(), conn.ConnectedClusterName(), conn.ConnectedServerId(), conn.Status().String(), conn.Servers(), conn.DiscoveredServers())
	}))
	sOption = append(sOption, nats.ErrorHandler(func(conn *nats.Conn, s *nats.Subscription, err error) {
		if s != nil {
			log.Printf("ERROR client connection async err status:%v servers:%v in %q/%q,err:%v\n", conn.Status().String(), conn.Servers(), s.Subject, s.Queue, err)
		} else {
			log.Printf("ERROR client connection async err status:%v servers:%v err:%v\n", conn.Status().String(), conn.Servers(), err)
		}
	}))

	tnc, err := nats.Connect(strNatsServer, sOption...)
	if err != nil {
		log.Printf(" %v 连接Nats服务器失败, 服务器:%v err:%v\n", AppName, strNatsServer, err)
		return nil
	}

	if isInitJetStream {
		tjs, jserr := tnc.JetStream(nats.PublishAsyncMaxPending(256))
		if jserr != nil {
			log.Printf(" %v 连接Nats服务器失败, 服务器:%v err:%v\n", AppName, strNatsServer, jserr)
			_ = tnc.Drain()
			return nil
		}
		log.Printf(" %v 连接Nats JetStream服务器成功, 服务器:%v\n", AppName, strNatsServer)
		tNatsObj := &NatsObj{
			nc:  tnc,
			jet: tjs,
		}
		return tNatsObj
	} else {
		log.Printf(" %v 连接Nats服务器成功, 服务器:%v\n", AppName, strNatsServer)
		tNatsObj := &NatsObj{
			nc:  tnc,
			jet: nil,
		}
		return tNatsObj
	}
}

// 将已收到缓冲区里的消息全部处理掉, 并关闭接收
func (sf *NatsObj) Drain() {
	if sf.nc != nil {
		_ = sf.nc.Drain()
	}
}

// 刷新
func (sf *NatsObj) Flush() error {
	e := sf.nc.Flush()
	return e
}

// 发布消息
func (sf *NatsObj) Publish(strSubj string, sData []byte) error {
	err := sf.nc.Publish(strSubj, sData)
	return err
}
func (sf *NatsObj) PublishMsg(tMsg *nats.Msg) error {
	err := sf.nc.PublishMsg(tMsg)
	return err
}

// 请求消息
func (sf *NatsObj) Request(strSubj string, sData []byte, tTimeout time.Duration) (*nats.Msg, error) {
	t, e := sf.nc.Request(strSubj, sData, tTimeout)
	return t, e
}

// 发布需响应的消息
func (sf *NatsObj) PublishRequest(strSubj, strReply string, sData []byte) error {
	err := sf.nc.PublishRequest(strSubj, strReply, sData)
	return err
}

// 异步订阅
func (sf *NatsObj) Subscribe(strSubj string, tCB nats.MsgHandler) (*nats.Subscription, error) {
	t, e := sf.nc.Subscribe(strSubj, tCB)
	return t, e
}

// 同步订阅
func (sf *NatsObj) SubscribeSync(strSubj string) (*nats.Subscription, error) {
	t, e := sf.nc.SubscribeSync(strSubj)
	return t, e
}

// 队列异步订阅
func (sf *NatsObj) QueueSubscribe(strSubj, strQueue string, tCb nats.MsgHandler) (*nats.Subscription, error) {
	t, e := sf.nc.QueueSubscribe(strSubj, strQueue, tCb)
	return t, e
}

// 队列同步订阅
func (sf *NatsObj) QueueSubscribeSync(strSubj, strQueue string) (*nats.Subscription, error) {
	t, e := sf.nc.QueueSubscribeSync(strSubj, strQueue)
	return t, e
}

// 消息管道订阅
// 下面使用例子
// ch := make(chan *nats.Msg, 64)
// sub, err := sf.nc.ChanSubscribe("foo", ch)
// msg := <- ch
func (sf *NatsObj) ChanSubscribe(strSubj string, tCh chan *nats.Msg) (*nats.Subscription, error) {
	t, e := sf.nc.ChanSubscribe(strSubj, tCh)
	return t, e
}

func (sf *NatsObj) JetPublish(strSubj string, sData []byte, sOpts ...nats.PubOpt) (*nats.PubAck, error) {
	t, e := sf.jet.Publish(strSubj, sData, sOpts...)
	return t, e
}

func (sf *NatsObj) JetPublishMsg(tMsg *nats.Msg, sOpts ...nats.PubOpt) (*nats.PubAck, error) {
	t, e := sf.jet.PublishMsg(tMsg, sOpts...)
	return t, e
}

func (sf *NatsObj) JetPublishMsgAsync(tMsg *nats.Msg, sOpts ...nats.PubOpt) (nats.PubAckFuture, error) {
	t, e := sf.jet.PublishMsgAsync(tMsg, sOpts...)
	return t, e
}

func (sf *NatsObj) JetPublishAsync(strSubj string, sData []byte, sOpts ...nats.PubOpt) (nats.PubAckFuture, error) {
	t, e := sf.jet.PublishAsync(strSubj, sData, sOpts...)
	return t, e
}

func (sf *NatsObj) JetSubscribe(strSubj string, tCB nats.MsgHandler, sOpts ...nats.SubOpt) (*nats.Subscription, error) {
	t, e := sf.jet.Subscribe(strSubj, tCB, sOpts...)
	return t, e
}

func (sf *NatsObj) JetSubscribeSync(strSubj string, sOpts ...nats.SubOpt) (*nats.Subscription, error) {
	t, e := sf.jet.SubscribeSync(strSubj, sOpts...)
	return t, e
}

func (sf *NatsObj) JetQueueSubscribe(strSubj, strQueue string, tCb nats.MsgHandler, sOpts ...nats.SubOpt) (*nats.Subscription, error) {
	t, e := sf.jet.QueueSubscribe(strSubj, strQueue, tCb, sOpts...)
	return t, e
}

func (sf *NatsObj) JetQueueSubscribeSync(strSubj, strQueue string, sOpts ...nats.SubOpt) (*nats.Subscription, error) {
	t, e := sf.jet.QueueSubscribeSync(strSubj, strQueue, sOpts...)
	return t, e
}

func (sf *NatsObj) JetPullSubscribe(strSubj, strDurable string, sOpts ...nats.SubOpt) (*nats.Subscription, error) {
	t, e := sf.jet.PullSubscribe(strSubj, strDurable, sOpts...)
	return t, e
}

func (sf *NatsObj) AddJetStream(tCfg *nats.StreamConfig, sOpts ...nats.JSOpt) (*nats.StreamInfo, error) {
	tInfo, _ := sf.jet.StreamInfo(tCfg.Name)
	if tInfo != nil {
		return tInfo, nil
	}

	i, e := sf.jet.AddStream(tCfg, sOpts...)
	return i, e
}
func (sf *NatsObj) UpdateJetStream(tCfg *nats.StreamConfig, sOpts ...nats.JSOpt) (*nats.StreamInfo, error) {
	i, e := sf.jet.UpdateStream(tCfg, sOpts...)
	return i, e
}
func (sf *NatsObj) DeleteJetStream(strStreamName string, sOpts ...nats.JSOpt) error {
	e := sf.jet.DeleteStream(strStreamName, sOpts...)
	return e
}

func (sf *NatsObj) AddJetConsumer(strStream string, tCfg *nats.ConsumerConfig, sOpts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	i, e := sf.jet.AddConsumer(strStream, tCfg, sOpts...)
	return i, e
}

func (sf *NatsObj) UpdateJetConsumer(strStream string, tCfg *nats.ConsumerConfig, sOpts ...nats.JSOpt) (*nats.ConsumerInfo, error) {
	i, e := sf.jet.UpdateConsumer(strStream, tCfg, sOpts...)
	return i, e
}

func (sf *NatsObj) DeleteJetConsumer(strStreamName, strConsumer string, sOpts ...nats.JSOpt) error {
	e := sf.jet.DeleteConsumer(strStreamName, strConsumer, sOpts...)
	return e
}
