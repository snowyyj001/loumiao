package message

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/log"
	"github.com/snowyyj001/loumiao/util"

	"github.com/golang/protobuf/proto"
)

//消息pack,unpack格式
//消息buffer以大端传输
//2+2+name+msg
//1，2字节代表消息报总长度
//3，4字节代表消息名长度
//name：消息名
//msg具体的消息包

var Encode func(name string, packet interface{}) ([]byte, int)
var Decode func(buff []byte, length int) (error, string, interface{})

func EncodeProBuff(name string, packet interface{}) ([]byte, int) {
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
	nLen := 4 + pLen + len(buff)

	binary.Write(bytesBuffer, binary.BigEndian, int16(nLen))
	binary.Write(bytesBuffer, binary.BigEndian, int16(pLen))
	binary.Write(bytesBuffer, binary.BigEndian, []byte(name))
	if buff != nil {
		binary.Write(bytesBuffer, binary.BigEndian, buff)
	}
	return bytesBuffer.Bytes(), nLen
}

func DecodeProBuff(buff []byte, length int) (error, string, interface{}) {
	mbuff1 := buff[2:4]
	nLen1 := util.BytesToUInt16(mbuff1, binary.BigEndian)
	if nLen1 <= 0 {
		log.Warning("Decode: packet name length is 0")
		return nil, "", nil
	}
	name := string(buff[4 : 4+nLen1])
	if length == 4+int(nLen1) { //just for on CONNECT/DISCONNECT
		return nil, name, nil
	}
	packet := GetPakcet(name)
	if packet == nil {
		log.Warningf("Decode: packet may not registered", name, nLen1, length)
		return nil, name, nil
	}
	err := proto.Unmarshal(buff[4+nLen1:length], packet.(proto.Message))
	if util.CheckErr(err) {
		return err, name, nil
	}
	return nil, name, packet
}

func EncodeJson(name string, packet interface{}) ([]byte, int) {
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
	nLen := 4 + pLen + len(buff)

	binary.Write(bytesBuffer, binary.BigEndian, int16(nLen))
	binary.Write(bytesBuffer, binary.BigEndian, int16(pLen))
	binary.Write(bytesBuffer, binary.BigEndian, []byte(name))
	if buff != nil {
		binary.Write(bytesBuffer, binary.BigEndian, buff)
	}
	return bytesBuffer.Bytes(), nLen
}

func DecodeJson(buff []byte, length int) (error, string, interface{}) {
	mbuff1 := buff[2:4]
	nLen1 := util.BytesToUInt16(mbuff1, binary.BigEndian)
	if nLen1 <= 0 {
		log.Warning("Decode: packet name length is 0")
		return nil, "", nil
	}
	name := string(buff[4 : 4+nLen1])
	if length == 4+int(nLen1) { //just for on CONNECT/DISCONNECT
		return nil, name, nil
	}
	packet := GetPakcet(name)
	if packet == nil {
		log.Warningf("Decode: packet may not registered", name, nLen1, length)
		return nil, name, nil
	}
	err := json.Unmarshal(buff[4+nLen1:length], packet)
	if util.CheckErr(err) {
		return err, name, nil
	}
	return nil, name, packet
}

func init() {
	if config.NET_PROTOCOL == "JSON" {
		Encode = EncodeJson
		Decode = DecodeJson
	} else {
		Encode = EncodeProBuff
		Decode = DecodeProBuff
	}
}
