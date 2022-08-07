package message

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/snowyyj001/loumiao/base"
	"github.com/snowyyj001/loumiao/config"

	"github.com/snowyyj001/loumiao/llog"
	"github.com/snowyyj001/loumiao/util"

	"github.com/golang/protobuf/proto"
)

//消息pack,unpack格式
//消息buffer以大端传输
//4+2+2+name+msg
//1，2，3，4字节代表消息报总长度
//5，6字节代表目标服务器类型/目标服务器uid
//7，8字节代表消息名长度

// Encode target: 目标服务器类型/id
//name: 消息名
//msg: 具体的消息包
var Encode func(target int, name string, packet interface{}) ([]byte, int)

// Decode uid: 服务器id
//buff: 消息包
//length: 包长度
var Decode func(uid int, buff []byte, length int) (error, int, string, interface{})

func EncodeProBuff(target int, name string, packet interface{}) ([]byte, int) {
	var pd proto.Message
	if name == "" {
		pd = packet.(proto.Message)
		name = GetMessageName(pd)
	}
	bytesBuffer := bytes.NewBuffer([]byte{})
	var buff []byte
	var err error
	if packet != nil {
		pd = packet.(proto.Message)
		buff, err = proto.Marshal(pd) //buff不为nil，是[]byte{},如果pd没有数据的话
	}
	if err != nil {
		llog.Errorf("EncodeProBuff: encode error: %s", err.Error())
		return nil, 0
	}

	pLen := len(name)
	nLen := 8 + pLen + len(buff)

	binary.Write(bytesBuffer, binary.BigEndian, int32(nLen))
	binary.Write(bytesBuffer, binary.BigEndian, int16(target))
	binary.Write(bytesBuffer, binary.BigEndian, int16(pLen))
	binary.Write(bytesBuffer, binary.BigEndian, []byte(name))

	if buff != nil {
		binary.Write(bytesBuffer, binary.BigEndian, buff)
	}
	if nLen > MaxPacketSize {
		llog.Errorf("EncodeProBuff: too big packet size: %d", nLen)
		return nil, 0
	}
	//llog.Debugf("EncodeProBuff: %s %d %d", name, nLen, bytesBuffer.Len())
	return bytesBuffer.Bytes(), nLen
}

func DecodeProBuff(uid int, buff []byte, length int) (error, int, string, interface{}) {
	mbuff1 := buff[4:6]
	target := int(base.BytesToUInt16(mbuff1, binary.BigEndian))

	if target > 0 && target != uid { //do not need decode anymore, a msg to other server
		return nil, target, "", nil
	}

	mbuff1 = buff[6:8]
	nameLen := base.BytesToUInt16(mbuff1, binary.BigEndian)
	if nameLen <= 0 {
		return fmt.Errorf("DecodeProBuff: msgname len is illegal: %d", nameLen), 0, "", nil
	}
	msgName := string(buff[8 : 8+nameLen])
	if length == 8+len(msgName) { //just for on CONNECT/DISCONNECT or []byte{}
		if filterWarning[msgName] {
			return nil, target, msgName, nil
		} else {
			packet := GetPakcet(msgName)
			if packet == nil {
				return fmt.Errorf("DecodeProBuff: packet[%s] may not registered, uid=%d,target=%d", msgName, uid, target), 0, "", nil
			}
			return nil, target, msgName, packet
		}

	}
	//fmt.Println("msgName = ", msgName)
	packet := GetPakcet(msgName)
	if packet == nil {
		return fmt.Errorf("DecodeProBuff: packet[%s] may not registered, uid=%d,target=%d", msgName, uid, target), 0, "", nil
	}
	err := proto.Unmarshal(buff[8+nameLen:length], packet.(proto.Message))
	if util.CheckErr(err) {
		llog.Errorf("DecodeProBuff: Unmarshal[%s] error", msgName)
		return err, target, "", nil
	}
	return nil, target, msgName, packet
}

func EncodeJson(target int, name string, packet interface{}) ([]byte, int) {
	if name == "" {
		name = GetMessageName(packet)
	}
	bytesBuffer := bytes.NewBuffer([]byte{})
	var buff []byte
	var err error

	if packet != nil {
		buff, err = json.Marshal(packet)
	}
	if util.CheckErr(err) {
		return nil, 0
	}

	pLen := len(name)
	nLen := 8 + pLen + len(buff)

	binary.Write(bytesBuffer, binary.BigEndian, int32(nLen))
	binary.Write(bytesBuffer, binary.BigEndian, int16(target))
	binary.Write(bytesBuffer, binary.BigEndian, int16(pLen))
	binary.Write(bytesBuffer, binary.BigEndian, []byte(name))
	if buff != nil {
		binary.Write(bytesBuffer, binary.BigEndian, buff)
	}
	return bytesBuffer.Bytes(), nLen
}

func DecodeJson(uid int, buff []byte, length int) (error, int, string, interface{}) {
	mbuff1 := buff[4:6]
	target := int(base.BytesToUInt16(mbuff1, binary.BigEndian))

	if target > 0 && target != uid { //do not need decode anymore, a msg to other server
		return nil, target, "", nil
	}

	mbuff1 = buff[6:8]
	nameLen := base.BytesToUInt16(mbuff1, binary.BigEndian)
	if nameLen <= 0 {
		return fmt.Errorf("DecodeJson: msgname len is illegal: %d", nameLen), 0, "", nil
	}
	msgName := string(buff[8 : 8+nameLen])
	if length == 8+int(nameLen) { //just for on CONNECT/DISCONNECT or []byte{}
		if filterWarning[msgName] {
			return nil, target, msgName, nil
		} else {
			return nil, target, msgName, GetPakcet(msgName)
		}
	}

	packet := GetPakcet(msgName)
	if packet == nil {
		return fmt.Errorf("DecodeJson: packet[%s] may not registered, uid=%d,target=%d", msgName, uid, target), 0, "", nil
	}
	err := json.Unmarshal(buff[8+nameLen:length], packet)
	if util.CheckErr(err) {
		llog.Errorf("DecodeJson: Unmarshal[%s] error", msgName)
		return err, target, "", nil
	}
	return nil, target, msgName, packet
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////
//

//对一个json或pb结构体序列化
var Pack func(packet interface{}) ([]byte, error)

//对buff反序列化为一个json或pb结构体
var UnPack func(packet interface{}, buff []byte) error

// UnPackHead 解析包头
//@ret: targetid, 消息名, 包体, error
func UnPackHead(buff []byte, length int) (int, string, []byte, error) {
	mbuff1 := buff[4:6]
	target := int(base.BytesToUInt16(mbuff1, binary.BigEndian))

	mbuff1 = buff[6:8]
	nameLen := base.BytesToUInt16(mbuff1, binary.BigEndian)
	if nameLen <= 0 {
		return 0, "", nil, fmt.Errorf("UnPackHead: msgname len is illegal: %d", nameLen)
	}
	msgName := string(buff[8 : 8+nameLen])
	return target, msgName, buff[8+nameLen : length], nil
}

func PackJson(packet interface{}) ([]byte, error) {
	return json.Marshal(packet)
}

func UnPackJson(packet interface{}, buff []byte) error {
	err := json.Unmarshal(buff, packet)
	if err != nil {
		llog.Errorf("message.UnPackJson: err = %s, buff = %s", err.Error(), string(buff))
	}
	return err
}

func PackProto(packet interface{}) ([]byte, error) {
	return proto.Marshal(packet.(proto.Message))
}

func UnPackProto(packet interface{}, buff []byte) error {
	err := proto.Unmarshal(buff, packet.(proto.Message))
	if err != nil {
		llog.Errorf("message.UnPackProto: packet = %v, err = %s", packet, err.Error())
	}
	return err
}

//udp不会粘包，所以不需要长度
type PacketUdp struct {
	UserId int64
	CmdId  uint16
}

func PackUdp(cmd uint16, userId int64, buffer []byte) []byte {
	var head PacketUdp
	buf := new(bytes.Buffer)
	head.CmdId = cmd
	head.UserId = userId
	binary.Write(buf, binary.BigEndian, head)
	buf.Write(buffer)
	return buf.Bytes()
}

func UpPackUdp(buffer []byte) (uint16, int64, []byte) {
	var head PacketUdp
	bytebuff := bytes.NewBuffer(buffer)
	binary.Read(bytebuff, binary.BigEndian, &head)
	return head.CmdId, head.UserId, bytebuff.Bytes()
}

func DoInit() {
	if config.NET_PROTOCOL == "JSON" {
		Encode = EncodeJson
		Decode = DecodeJson
		Pack = PackJson
		UnPack = UnPackJson
	} else {
		Encode = EncodeProBuff
		Decode = DecodeProBuff
		Pack = PackProto
		UnPack = UnPackProto
	}
}
