package util

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/snowyyj001/loumiao/llog"
	"math"
	rand2 "math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand2.Seed(time.Now().UnixNano())
}

func Recover() {
	if r := recover(); r != nil {
		buf := make([]byte, 2048)
		l := runtime.Stack(buf, false)
		llog.Errorf("Recover %v: %s", r, buf[:l])
	}
}

// 获得rpc注册名
func RpcFuncName(call interface{}) string {
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	funcName := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	return funcName
}

//随机数[0,n)
func Random(n int) int {
	return int(rand2.Int31n(int32(n)))
}

//随机数[n1,n2)
func Randomd(n1, n2 int) int {
	return n1 + int(rand2.Int31n(int32(n2-n1)))
}

//随机数[0,n)
func Random64(n int64) int64 {
	return rand2.Int63n(n)
}

//随机数[n1,n2)
func Randomd64(n1, n2 int64) int64 {
	return n1 + rand2.Int63n(n2-n1)
}

// Clamp 截断
func Clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

//当前格式化时间字符串
func TimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

//当前格式化时间字符串
func TimeStrFormat(mat string) string {
	return time.Now().Format(mat)
}

//当前时间
func NowTime() time.Time {
	return time.Now()
}

//时间戳秒
func TimeStampSec() int64 {
	return time.Now().Unix()
}

//时间戳毫秒
func TimeStamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//指定日期的时间戳毫秒
func TimeStampTarget(y int, m time.Month, d int, h int, mt int, s int) int64 {
	return time.Date(y, m, d, h, mt, s, 0, time.Local).UnixNano() / int64(time.Millisecond)
}

func Assert(x bool, y string) {
	if x == false {
		llog.Fatalf("util.Assert: %s", y)
	}
}

func CheckErr(err error) bool {
	if err != nil {
		return true
	}
	return false
}

func FloorInt(v float64) int {
	nv := math.Floor(v)
	return int(nv)
}

func FloorInt64(v float64) int64 {
	nv := math.Floor(v)
	return int64(nv)
}

func Max(a, b int) int { //官方包里提供了float64的max和min函数，原因就是，golang不支持泛型，非要提供会让代码不够优雅的
	if a > b {
		return a
	}
	return b
}

func Max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Min64(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

func CopyArray(src []int, size int) []int {
	dst := make([]int, size, size)
	copy(dst, src)
	return dst
}

func RemoveSliceString(arr []string, val string) []string {
	for i, v := range arr {
		if v == val {
			arr = append(arr[:i], arr[i+1:]...)
			return arr
		}
	}
	return arr
}

func RemoveSliceByIndex(arr []int, index int) []int {
	arr = append(arr[:index], arr[index+1:]...)
	return arr
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

func RemoveSliceArray(dst []int, src []int, size int) []int {
	for i := 0; i < size; i++ {
		dst = RemoveSlice(dst, src[i])
	}
	return dst
}

func EncodeBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func DecodeBase64(str string) string {
	if ret, err := base64.StdEncoding.DecodeString(str); err == nil {
		return string(ret)
	}
	return ""
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
		llog.Errorf("Atoi strconv.Atoi failed : num = %s, err = %s", num, err.Error())
	}
	return val
}

func Atoi64(num string) int64 {
	val, err := strconv.Atoi(num)
	if err != nil {
		llog.Errorf("Atoi strconv.Atoi64 failed : num = %s, err = %s", num, err.Error())
	}
	return int64(val)
}

func Itoa(num int) string {
	return strconv.Itoa(num)
}

func Itoa64(num int64) string {
	return strconv.Itoa(int(num))
}

func Array2String(arr []int) string {
	if len(arr) == 0 {
		return ""
	}
	var sb strings.Builder
	sz := len(arr)
	for i := 0; i < sz; i++ {
		sb.WriteString(strconv.Itoa(arr[i]))
		if i < sz-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func String2Array(str string) []int {
	arrstr := strings.Split(str, ",")
	sz := len(arrstr)
	arr := make([]int, sz)
	for i := 0; i < sz; i++ {
		arr[i], _ = strconv.Atoi(arrstr[i])
	}
	return arr
}

//@version: format 1.0.1,max value is 999
func GetVersionCode(version string) int {
	arr := strings.Split(version, ".")
	vcode := Atoi(arr[0])*1000*1000 + Atoi(arr[1])*1000 + Atoi(arr[2])
	return vcode
}

func GetVersionString(version int) string {
	v1 := version / (1000 * 1000)
	v2 := version / 1000 % 1000
	v3 := version % 1000

	return fmt.Sprintf("%d.%d.%d", v1, v2, v3)
}
