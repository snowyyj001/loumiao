package message

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"

	"github.com/golang/protobuf/proto"
)

//消息pack,unpack格式
//消息buffer以大端传输
//2+4+4+2+name+msg
//1，2字节代表消息报总长度
//3，4，5，6字节代表目标服务器id
//7，8，9，10字节保留字段
//11，12字节代表消息名长度

//target: 目标服务器id
//flag: 服务器标识参数
//name: 消息名
//msg: 具体的消息包
var Encode func(target int, flag int, name string, packet interface{}) ([]byte, int)

//uid: 服务器id
//buff: 消息包
//length: 包长度
var Decode func(uid int, buff []byte, length int) (error, int, string, interface{})

func EncodeProBuff(target int, flag int, name string, packet interface{}) ([]byte, int) {
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
		buff, err = proto.Marshal(pd)
	}
	if util.CheckErr(err) {
		return nil, 0
	}

	pLen := len(name)
	nLen := 12 + pLen + len(buff)

	binary.Write(bytesBuffer, binary.BigEndian, int16(nLen))
	binary.Write(bytesBuffer, binary.BigEndian, int32(target))
	binary.Write(bytesBuffer, binary.BigEndian, int32(flag))
	binary.Write(bytesBuffer, binary.BigEndian, int16(pLen))
	binary.Write(bytesBuffer, binary.BigEndian, []byte(name))

	if buff != nil {
		binary.Write(bytesBuffer, binary.BigEndian, buff)
	}
	return bytesBuffer.Bytes(), nLen
}

func DecodeProBuff(uid int, buff []byte, length int) (error, int, string, interface{}) {
	mbuff1 := buff[2:6]
	target := int(util.BytesToUInt32(mbuff1, binary.BigEndian))

	if target != -1 && target != uid { //do not need decode anymore
		return nil, target, "", nil
	}
	mbuff1 = buff[6:10]
	flag := util.BytesToUInt32(mbuff1, binary.BigEndian)
	if flag <= 0 {
		return fmt.Errorf("DecodeProBuff: flag is illegal: %d", flag), 0, "", nil
	}

	mbuff1 = buff[10:12]
	nameLen := util.BytesToUInt16(mbuff1, binary.BigEndian)
	if nameLen <= 0 {
		return fmt.Errorf("DecodeProBuff: msgname len is illegal: %d", nameLen), 0, "", nil
	}

	msgName := string(buff[12 : 12+nameLen])
	if length == 12+int(nameLen) { //just for on CONNECT/DISCONNECT
		return nil, target, msgName, nil
	}
	packet := GetPakcet(msgName)
	if packet == nil {
		return fmt.Errorf("DecodeProBuff: packet[%s] may not registered", msgName), 0, "", nil
	}
	err := proto.Unmarshal(buff[12+nameLen:length], packet.(proto.Message))
	if util.CheckErr(err) {
		log.Warningf("DecodeProBuff: Unmarshal[%s] error", msgName)
		return err, target, "", nil
	}
	return nil, target, msgName, packet
}

func EncodeJson(target int, flag int, name string, packet interface{}) ([]byte, int) {
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
	nLen := 12 + pLen + len(buff)

	binary.Write(bytesBuffer, binary.BigEndian, int16(nLen))
	binary.Write(bytesBuffer, binary.BigEndian, int16(pLen))
	binary.Write(bytesBuffer, binary.BigEndian, int32(target))
	binary.Write(bytesBuffer, binary.BigEndian, int32(flag))
	binary.Write(bytesBuffer, binary.BigEndian, []byte(name))
	if buff != nil {
		binary.Write(bytesBuffer, binary.BigEndian, buff)
	}
	return bytesBuffer.Bytes(), nLen
}

func DecodeJson(uid int, buff []byte, length int) (error, int, string, interface{}) {
	mbuff1 := buff[2:6]
	target := int(util.BytesToUInt32(mbuff1, binary.BigEndian))
	if target <= 0 {
		return fmt.Errorf("DecodeJson: target is illegal: %d", target), 0, "", nil
	}

	if target != -1 && target != uid { //do not need decode anymore
		return nil, target, "", nil
	}

	mbuff1 = buff[6:10]
	flag := util.BytesToUInt32(mbuff1, binary.BigEndian)
	if flag <= 0 {
		return fmt.Errorf("DecodeJson: flag is illegal: %d", flag), 0, "", nil
	}
	mbuff1 = buff[10:12]
	nameLen := util.BytesToUInt16(mbuff1, binary.BigEndian)
	if nameLen <= 0 {
		return fmt.Errorf("DecodeJson: msgname len is illegal: %d", nameLen), 0, "", nil
	}
	msgName := string(buff[12 : 12+nameLen])
	if length == 12+int(nameLen) { //just for on CONNECT/DISCONNECT
		return nil, target, msgName, nil
	}

	packet := GetPakcet(msgName)
	if packet == nil {
		return fmt.Errorf("DecodeJson: packet[%s] may not registered", msgName), 0, "", nil
	}
	err := json.Unmarshal(buff[12+nameLen:length], packet)
	if util.CheckErr(err) {
		log.Warningf("DecodeJson: Unmarshal[%s] error", msgName)
		return err, target, "", nil
	}
	return nil, target, msgName, packet
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
