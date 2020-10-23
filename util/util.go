package util

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
	rand2 "math/rand"
	"strconv"
	"time"

	"github.com/snowyyj001/loumiao/log"
)

func init() {
	rand2.Seed(time.Now().UnixNano())
}

//随机数[0,n)
func Random(n int) int {
	return int(rand2.Int31n(int32(n)))
}

//时间戳毫秒
func TimeStamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//指定日期的时间戳毫秒
func TimeStampTarget(y int, m time.Month, d int, h int, mt int, s int) int64 {
	return time.Date(y, m, d, h, mt, s, 0, time.Local).UnixNano() / int64(time.Millisecond)
}

func CheckErr(err error) bool {
	if err != nil {
		log.Error("CheckErr: " + err.Error())
		return true
	}
	return false
}

func BytesToUInt16(buff []byte, order binary.ByteOrder) uint16 {
	bytebuff := bytes.NewBuffer(buff)
	var data uint16
	binary.Read(bytebuff, order, &data)
	return data
}

func BytesToInt16(buff []byte, order binary.ByteOrder) int16 {
	bytebuff := bytes.NewBuffer(buff)
	var data int16
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

func FloorInt(v int) int {
	nv := math.Floor(float64(v))
	return int(nv)
}

func FloorInt64(v int64) int64 {
	nv := math.Floor(float64(v))
	return int64(nv)
}

func CopyArray(dst []int, src []int, size int) {
	for i := 0; i < size; i++ {
		dst[i] = src[i]
	}
}

func RemoveSlice(arr []int, val int) []int {
	for i, v := range arr {
		if v == val {
			arr = append(arr[:i], arr[i+1:]...)
			return arr
		}
	}
	return arr
}

func Md5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func Atoi(num string) int {
	val, err := strconv.Atoi(num)
	if err != nil {
		log.Fatal("Atoi strconv.Atoi failed " + err.Error())
	}
	return val
}

func Itoa(num int) string {
	return strconv.Itoa(num)
}

func Itoa64(num int64) string {
	return strconv.Itoa(int(num))
}

func HasBit(val int, flag int) bool {
	return (val & flag) != 0
}

func EqualBit(val int, flag int) bool {
	return (val & flag) == flag
}
