package lnats

import (
	"fmt"
	"github.com/prometheus/common/log"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/config"
)

var (
	lnc *nats.Conn
)

func FormatTopic(topic, prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, topic)
}

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

//请求消息
func Request(topic string, message []byte) []byte {
	msg, err := lnc.Request(topic, message, 3*time.Second)
	if err != nil {
		return []byte{}
	}
	return msg.Data
}

//回复消息
func Response(topic string, call func([]byte) []byte) error {
	_, err := lnc.Subscribe(topic, func(m *nats.Msg) {
		data := call(m.Data)
		m.Respond(data)
	})
	return err
}

//请求消息
func RequestTag(topic string, prefix string, message []byte) []byte {
	newtopic := FormatTopic(topic, prefix)
	msg, err := lnc.Request(newtopic, message, 3*time.Second)
	if err != nil {
		return []byte{}
	}
	return msg.Data
}

//回复消息
func ResponseTag(topic string, prefix string, call func([]byte) []byte) error {
	newtopic := FormatTopic(topic, prefix)
	_, err := lnc.Subscribe(newtopic, func(m *nats.Msg) {
		data := call(m.Data)
		m.Respond(data)
	})
	return err
}

func Init(addr []string) {
	target := strings.Join(addr, ",")
	name := nats.Name(config.NET_GATE_SADDR)

	nc, err := nats.Connect(target, name,
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Errorf("NATS client connection got disconnected: %s %s", nc.LastError(), err.Error())
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Errorf("NATS client reconnected after a previous disconnection, connected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Errorf("NATS client connection closed: %s", nc.LastError())
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Errorf("NATS client on %s encountered an error: %s", nc.ConnectedUrl(), err.Error())
		}))

	if err != nil {
		log.Fatalf("nats connect failed: %s", target)
	} else {
		log.Infof("nats connect success: nickname=%s,target=%s", config.NET_GATE_SADDR, target)
	}
	lnc = nc
}

func init() {

}
