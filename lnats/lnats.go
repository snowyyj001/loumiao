package lnats

import (
	"encoding/json"
	"fmt"
	"github.com/snowyyj001/loumiao/define"
	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/util"
	"github.com/snowyyj001/loumiao/util/ratelimiter"
	"strings"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/config"
)

var (
	lnc *nats.Conn
	errorLimiter *ratelimiter.RateLimiter			//上报error信息限流
)

const (
	TIMEOUT_NATS = 3

	PermitsPerSecond = 10		//100ms 一个
	MaxPermits = 1024
)

func FormatTopic(topic, prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, topic)
}

//不带分组得消息订阅发布，会采用one-many的方式
//带分组的方式，会采用one-one的方式

//同步订阅消息,带分组
func QueueSubscribeTagSync(topic, prefix, queue string) ([]byte, error) {
	newtopic := FormatTopic(topic, prefix)
	return QueueSubscribeSync(newtopic, queue, TIMEOUT_NATS)
}

//同步订阅消息,带分组
func QueueSubscribeSync(topic, queue string, waittime int) ([]byte, error) {
	// Subscribe
	sub, err := lnc.QueueSubscribeSync(topic, queue)
	if err != nil {
		return nil, err
	}
	// Wait for a message
	msg, err := sub.NextMsg(time.Duration(waittime) * time.Second)
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

//异步订阅消息,带分组
func QueueSubscribeTag(topic, prefix, queue string, call func([]byte)) error {
	newtopic := FormatTopic(topic, prefix)
	return QueueSubscribe(newtopic, queue, call)
}

//同步订阅消息
func SubscribeTagSync(topic string, prefix string) ([]byte, error) {
	newtopic := FormatTopic(topic, prefix)
	return SubscribeSync(newtopic, TIMEOUT_NATS)
}

//同步订阅消息
func SubscribeSync(topic string, waittime int) ([]byte, error) {
	// Subscribe
	sub, err := lnc.SubscribeSync(topic)
	if err != nil {
		return nil, err
	}
	// Wait for a message
	msg, err := sub.NextMsg(time.Duration(waittime) * time.Second)
	if err != nil {
		return nil, err
	}
	return msg.Data, err
}

//异步订阅消息
func SubscribeTagAsyn(topic string, prefix string, call func([]byte)) error {
	newtopic := FormatTopic(topic, prefix)
	return SubscribeAsyn(newtopic, call)
}

//异步订阅消息
func SubscribeAsyn(topic string, call func([]byte)) error {
	// Subscribe
	_, err := lnc.Subscribe(topic, func(m *nats.Msg) {
		defer util.Recover()
		call(m.Data)
	})
	return err
}

//发布消息
func PublishTag(topic string, prefix string, message []byte) error {
	newtopic := FormatTopic(topic, prefix)
	return lnc.Publish(newtopic, message)
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
func Request(topic string, message []byte, waittime int) []byte {
	msg, err := lnc.Request(topic, message, time.Duration(waittime)*time.Second)
	if err != nil {
		return []byte(err.Error())
	}
	return msg.Data
}

//回复消息
func Response(topic string, call func([]byte) []byte) error {
	_, err := lnc.Subscribe(topic, func(m *nats.Msg) {
		defer util.Recover()
		data := call(m.Data)
		m.Respond(data)
	})
	return err
}

//请求消息
func RequestTag(topic string, prefix string, message []byte, waittime int) []byte {
	newtopic := FormatTopic(topic, prefix)
	msg, err := lnc.Request(newtopic, message, time.Duration(waittime)*time.Second)
	if err != nil {
		return []byte(err.Error())
	}
	return msg.Data
}

//回复消息
func ResponseTag(topic string, prefix string, call func([]byte) []byte) error {
	newtopic := FormatTopic(topic, prefix)
	_, err := lnc.Subscribe(newtopic, func(m *nats.Msg) {
		defer util.Recover()
		data := call(m.Data)
		m.Respond(data)
	})
	return err
}

//回复消息带分组
func QueueResponse(topic string, queue string, call func([]byte) []byte) error {
	_, err := lnc.QueueSubscribe(topic, queue, func(m *nats.Msg) {
		defer util.Recover()
		data := call(m.Data)
		m.Respond(data)
	})
	return err
}

//回复消息带分组
func QueueResponseTag(topic string, prefix string, queue string, call func([]byte) []byte) error {
	newtopic := FormatTopic(topic, prefix)
	_, err := lnc.QueueSubscribe(newtopic, queue, func(m *nats.Msg) {
		defer util.Recover()
		data := call(m.Data)
		m.Respond(data)
	})
	return err
}

func Init() {
	target := strings.Join(config.Cfg.NatsAddr, ",")
	name := nats.Name(config.NET_GATE_SADDR)

	nc, err := nats.Connect(target, name,
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			str := fmt.Sprintf("NATS client connection got disconnected: %s %s", nc.LastError(), err.Error())
			llog.Error(str)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			str := fmt.Sprintf("NATS client reconnected after a previous disconnection, connected to %s", nc.ConnectedUrl())
			llog.Error(str)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			str := fmt.Sprintf("NATS client connection closed: %s", nc.LastError())
			llog.Error(str)
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			str := fmt.Sprintf("NATS client on %s encountered an error: %s", nc.ConnectedUrl(), err.Error())
			if err == nats.ErrSlowConsumer {
				pendingMsgs, _, err := sub.Pending()
				if err != nil {
					llog.Errorf("couldn't get pending messages: %s", err.Error())
					return
				}
				str = fmt.Sprintf("%s\nFalling behind with %d pending messages on subject %q.\n", str, pendingMsgs, sub.Subject)
				// Log error, notify operations...
			}
			llog.Error(str)
		}))

	if err != nil {
		str := fmt.Sprintf("nats connect failed: %s", target)
		llog.Fatal(str)
	} else {
		str := fmt.Sprintf("nats connect success: %s", target)
		llog.Info(str)
	}
	lnc = nc

	errorLimiter = ratelimiter.Create(PermitsPerSecond, MaxPermits)
}

//上报服务器关键信息
func ReportMail(tag int, str string) {
	//这里说一下golang有意思的事情，errorLimiter为nil的话，是不会在241行报错的，会在RateLimiter的Acquire内部使用self时报错，这和其他语言例如cpp是不一样的
	if errorLimiter == nil {
		return
	}
	errorLimiter.Acquire()
	reqParam := &struct {
		Tag     int    `json:"tag"`     //邮件类型
		Id      int    `json:"id"`      //区服id
		Content string `json:"content"` //邮件内容
	}{}
	reqParam.Tag = tag
	reqParam.Id = config.NET_NODE_ID
	reqParam.Content = fmt.Sprintf("uid: %d \nname: %s\nhost: %s\r\ncontent: %s", config.SERVER_NODE_UID, config.SERVER_NAME, config.NET_GATE_SADDR, str)
	buffer, err := json.Marshal(&reqParam)
	if err == nil {
		Publish(define.TOPIC_SERVER_MAIL, buffer)
	}
}

func init() {
	llog.ReportMailHanlder = ReportMail
}
