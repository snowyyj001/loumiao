package nsq

import (
	"fmt"
	"github.com/snowyyj001/loumiao/log"
	"time"

	"github.com/snowyyj001/loumiao/util"

	"github.com/nsqio/go-nsq"
)

var nsqCfg *nsq.Config
var nsqdAddress []string
var producer *nsq.Producer

// 消费者
type Consumer interface {
	HandleMessage(msg *nsq.Message) error
}

//consumer
func RegisterTpoic(topic string, channel string, recv Consumer) {
	c, err := nsq.NewConsumer(topic, channel, nsqCfg) // 新建一个消费者
	if err != nil {
		panic(err)
	}
	c.SetLogger(nil, 0) //屏蔽系统日志
	//AddConurrentHandlers
	c.AddHandler(recv) // 添加消费者接口

	//建立NSQLookupd连接
	if err := c.ConnectToNSQDs(nsqdAddress); err != nil {
	//if err := c.ConnectToNSQLookupd("127.0.0.1:4160"); err != nil {
		panic(err)
	}
}

// 初始化生产者
func InitProducer(str string) {
	var err error
	log.Debugf("InitProducer addr: ", str)
	producer, err = nsq.NewProducer(str, nsq.NewConfig())
	if err != nil {
		panic(err)
	}
}

//发布消息
func Publish(topic string, message []byte) error {
	var err error
	if producer != nil {
		if len(message) == 0 {
			return nil
		}
		trytimes := 0
		for err = producer.Publish(topic, message); err != nil; err = producer.Publish(topic, message) {
			log.Error(Publish msg error)
			trytimes++
			addr := nsqdAddress[util.Random(len(nsqdAddress))]
			InitProducer(addr)
			if trytimes >= len(nsqdAddress) {
				break
			}
		}
		return err
	}
	return fmt.Errorf("producer is nil", err)
}

func Init(addr []string) {
	nsqdAddress = addr
	str := nsqdAddress[util.Random(len(nsqdAddress))]
	InitProducer(str)
}

func init() {
	nsqCfg := nsq.NewConfig()
	nsqCfg.LookupdPollInterval = time.Second //设置重连时间
}
