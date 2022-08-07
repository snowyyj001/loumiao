package base

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
	"math/rand"
	"strconv"
)

func Assert(x bool, y string) {
	if bool(x) == false {
		log.Println("Assert: %s", y)
	}
}

//随机数[i,n]
func RandI(i int, n int) int {
	if i > n {
		return i
	}
	return int(i + rand.Int()%(n-i+1))
}

//随机数[i,n]
func RandF(i float32, n float32) float32 {
	if i > n {
		return i
	}
	return i + (n-i)*rand.Float32()
}

//-----------string strconv type-------------//
func Int(str string) int {
	n, _ := strconv.Atoi(str)
	return n
}

func Int64(str string) int64 {
	n, _ := strconv.ParseInt(str, 0, 64)
	return n
}

func Float32(str string) float32 {
	n, _ := strconv.ParseFloat(str, 32)
	return float32(n)
}

func Float64(str string) float64 {
	n, _ := strconv.ParseFloat(str, 64)
	return n
}

func Bool(str string) bool {
	n, _ := strconv.ParseBool(str)
	return n
}

// IntToBytes 整形转换成字节
func IntToBytes(val int) []byte {
	tmp := uint32(val)
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, tmp)
	return buff
}

// BytesToInt 字节转换成整形
func BytesToInt(data []byte) int {
	buff := make([]byte, 4)
	copy(buff, data)
	tmp := int32(binary.LittleEndian.Uint32(buff))
	return int(tmp)
}

// Int16ToBytesDefault 整形16转换成字节
func Int16ToBytesDefault(val int16) []byte {
	tmp := uint16(val)
	buff := make([]byte, 2)
	binary.LittleEndian.PutUint16(buff, tmp)
	return buff
}

// BytesToInt16Default 字节转换成为int16
func BytesToInt16Default(data []byte) int16 {
	buff := make([]byte, 2)
	copy(buff, data)
	tmp := binary.LittleEndian.Uint16(buff)
	return int16(tmp)
}

// Int64ToBytesDefault 转化64位
func Int64ToBytesDefault(val int64) []byte {
	tmp := uint64(val)
	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, tmp)
	return buff
}

// BytesToInt64Default 转化64位
func BytesToInt64Default(data []byte) int64 {
	bytebuff := bytes.NewBuffer(data)
	var r int64
	binary.Read(bytebuff, binary.LittleEndian, &r)
	return r
}

//转化float
func Float32ToByte(val float32) []byte {
	tmp := math.Float32bits(val)
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, tmp)
	return buff
}

func BytesToFloat32(data []byte) float32 {
	buff := make([]byte, 4)
	copy(buff, data)
	tmp := binary.LittleEndian.Uint32(buff)
	return math.Float32frombits(tmp)
}

// Float64ToByte 转化float64
func Float64ToByte(val float64) []byte {
	tmp := math.Float64bits(val)
	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, tmp)
	return buff
}

func BytesToFloat64(data []byte) float64 {
	buff := make([]byte, 8)
	copy(buff, data)
	tmp := binary.LittleEndian.Uint64(buff)
	return math.Float64frombits(tmp)
}

func BytesToUInt16(buff []byte, order binary.ByteOrder) uint16 {
	bytebuff := bytes.NewBuffer(buff)
	var data uint16
	binary.Read(bytebuff, order, &data)
	return data
}

func BytesToUInt32(buff []byte, order binary.ByteOrder) uint32 {
	bytebuff := bytes.NewBuffer(buff)
	var data uint32
	binary.Read(bytebuff, order, &data)
	return data
}

func BytesToInt32(buff []byte, order binary.ByteOrder) int32 {
	bytebuff := bytes.NewBuffer(buff)
	var data int32
	binary.Read(bytebuff, order, &data)
	return data
}

// Int64ToBytes 转化64位
func Int64ToBytes(val int64, order binary.ByteOrder) []byte {
	tmp := uint64(val)
	buff := make([]byte, 8)
	order.PutUint64(buff, tmp)
	return buff
}

func BytesToInt64(data []byte, order binary.ByteOrder) int64 {
	bytebuff := bytes.NewBuffer(data)
	var r int64
	binary.Read(bytebuff, order, &r)
	return r
}
