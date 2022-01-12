package util

import (
	"crypto/md5"
	"fmt"
	"math"
	rand2 "math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/snowyyj001/loumiao/llog"
)

func init() {
	rand2.Seed(time.Now().UnixNano())
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

func FloorInt(v int) int {
	nv := math.Floor(float64(v))
	return int(nv)
}

func FloorInt64(v int64) int64 {
	nv := math.Floor(float64(v))
	return int64(nv)
}

func Max(a, b int) int {
	m := int(math.Max(float64(a), float64(b)))
	return m
}

func Max64(a, b int64) int64 {
	m := int64(math.Max(float64(a), float64(b)))
	return m
}

func Min(a, b int) int {
	m := int(math.Min(float64(a), float64(b)))
	return m
}

func Min64(a, b int64) int64 {
	m := int64(math.Min(float64(a), float64(b)))
	return m
}

func CopyArray(dst []int, src []int, size int) {
	for i := 0; i < size; i++ {
		dst[i] = src[i]
	}
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
		llog.Errorf("Atoi strconv.Atoi failed " + err.Error())
	}
	return val
}

func Atoi64(num string) int64 {
	val, err := strconv.Atoi(num)
	if err != nil {
		llog.Errorf("Atoi strconv.Atoi64 failed " + err.Error())
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
