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
//2+4+2+name+msg
//1，2字节代表消息报总长度
//3，4，5，6字节代表目标服务器id
//7，8字节代表消息名长度

//target: 目标服务器id
//name: 消息名
//msg: 具体的消息包
var Encode func(target int, name string, packet interface{}) ([]byte, int)

//uid: 服务器id
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
	if length == 8+int(nameLen) { //just for on CONNECT/DISCONNECT or []byte{}
		if filterWarning[msgName] {
			return nil, target, msgName, nil
		} else {
			return nil, target, msgName, GetPakcet(msgName)
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

func Pack(packet interface{}) ([]byte, error) {
	return proto.Marshal(packet.(proto.Message))
}

func UnPack(msgName string, buff []byte) (interface{}, error) {
	packet := GetPakcet(msgName)
	err := proto.Unmarshal(buff, packet.(proto.Message))
	return packet, err
}

func DoInit() {
	if config.NET_PROTOCOL == "JSON" {
		Encode = EncodeJson
		Decode = DecodeJson
	} else {
		Encode = EncodeProBuff
		Decode = DecodeProBuff
	}
}
