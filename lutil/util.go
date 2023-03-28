package lutil

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/snowyyj001/loumiao/llog"
	"math"
	rand2 "math/rand"
	"os"
	"os/exec"
	"path/filepath"
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
		llog.Errorf("Recover %v: %s", r, string(buf[:l]))
	}
}

func Go(fn func()) {
	go func() {
		defer Recover()
		fn()
	}()
}

// 获得rpc注册名
func RpcFuncName(call interface{}) string {
	//base64str := base64.StdEncoding.EncodeToString([]byte(funcName))
	funcName := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	return funcName
}

// 随机数[0,n)
func Random(n int) int {
	return int(rand2.Int31n(int32(n)))
}

// 随机数[n1,n2)
func Randomd(n1, n2 int) int {
	return n1 + int(rand2.Int31n(int32(n2-n1)))
}

// 随机数[0,n)
func Random64(n int64) int64 {
	return rand2.Int63n(n)
}

// 随机数[n1,n2)
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

// @version: format 1.0.1,max value is 999
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

func GetExeName() string {
	filePath, _ := exec.LookPath(os.Args[0])
	//llog.Println("os.Args ", os.Args)
	fileName := filepath.Base(filePath)
	if len(os.Args) > 1 {
		arg1 := os.Args[1]
		if len(arg1) > 0 {
			arg1 = strings.Replace(arg1, ":", "[", -1)
			fileName += "_" + arg1 + "]"
		}
	}

	return fileName
}

func DumpPid() {
	pid := os.Getpid()
	llog.Infof("pid: %d", pid)

	fileName := GetExeName()

	pidFile, err := os.OpenFile(fileName+".pid", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		llog.Errorf("OpenFile err: %s", err.Error())
	}
	pidFile.WriteString(fmt.Sprint(pid))

	pidFile.Close()
}
