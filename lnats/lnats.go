package lnats

import (
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/log"
)

var (
	lnc *nats.Conn
)

//不带分组得消息订阅发布，会采用one-many的方式
//带分组的方式，会采用one-one的方式
//同步订阅消息,带分组
func QueueSubscribeSync(topic, queue string) ([]byte, error) {
	// Subscribe
	sub, err := lnc.QueueSubscribeSync(topic, queue)
	if err != nil {
		return nil, err
	}
	// Wait for a message
	msg, err := sub.NextMsg(10 * time.Second)
	if err != nil {
		return nil, err
	}
	return msg.Data, err
}

//异步订阅消息,带分组
func QueueSubscribe(topic, queue string, call func([]byte)) error {
	// Subscribe
	_, err := lnc.QueueSubscribe(topic, queue, func(m *nats.Msg) {
		call(m.Data)
	})
	return err
}

//同步订阅消息
func SubscribeSync(topic string) ([]byte, error) {
	// Subscribe
	sub, err := lnc.SubscribeSync(topic)
	if err != nil {
		return nil, err
	}
	// Wait for a message
	msg, err := sub.NextMsg(10 * time.Second)
	if err != nil {
		return nil, err
	}
	return msg.Data, err
}

//异步订阅消息
func SubscribeAsyn(topic string, call func([]byte)) error {
	// Subscribe
	_, err := lnc.Subscribe(topic, func(m *nats.Msg) {
		call(m.Data)
	})
	return err
}

//发布消息
func Publish(topic string, message []byte) error {
	return lnc.Publish(topic, message)
}

//发布消息
func PublishString(topic, message string) error {
	return lnc.Publish(topic, []byte(message))
}

//发布消息
func PublishInt(topic string, message int) error {
	return lnc.Publish(topic, base.Int64ToBytes(int64(message)))
}

func Init(addr []string) {
	target := strings.Join(addr, ",")
	name := nats.Name(config.NET_GATE_SADDR)
	nc, err := nats.Connect(target, name)
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Infof("nats connect success: nickname=%s,target=%s", config.NET_GATE_SADDR, target)
	}
	lnc = nc
}

func init() {

}
