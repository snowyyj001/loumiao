package llog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/snowyyj001/loumiao/define"

	"github.com/snowyyj001/loumiao/lnats"

	go_logger "github.com/phachon/go-logger"
	"github.com/snowyyj001/loumiao/config"
)

var (
	logger *go_logger.Logger
)

func init() {
	logger = go_logger.NewLogger()
	logger.Detach("console")
	os.Mkdir("logs", os.ModePerm)
	os.Mkdir(fmt.Sprintf("logs/%s", config.SERVER_NAME), os.ModePerm)
	if config.GAME_LOG_CONLOSE {
		// 命令行输出配置
		consoleConfig := &go_logger.ConsoleConfig{
			Color:      true,                 // 命令行输出字符串是否显示颜色
			JsonFormat: config.GAME_LOG_JSON, // 命令行输出字符串是否格式化
			Format:     "",                   // 如果输出的不是 json 字符串，JsonFormat: false, 自定义输出的格式
		}
		// 添加 console 为 logger 的一个输出
		logger.Attach("console", config.GAME_LOG_LEVEL, consoleConfig)
	} else {
		filename := fmt.Sprintf("./logs/%s/%s.log", config.SERVER_NAME, config.SERVER_NAME)
		//filename := fmt.Sprintf("./logs/%s.%s.llog", config.SERVER_NAME, time.Now().Format("2006-01-02.15.04.05"))
		dealLogs(filename)
		// 文件输出配置
		fileConfig := &go_logger.FileConfig{
			Filename:   filename,             // 日志输出文件名，不自动存在
			MaxSize:    50 * 1024,            // 文件最大值（KB），默认值0不限
			MaxLine:    0,                    // 文件最大行数，默认 0 不限制
			DateSlice:  "d",                  // 文件根据日期切分， 支持 "Y" (年), "m" (月), "d" (日), "H" (时), 默认 "no"， 不切分
			JsonFormat: config.GAME_LOG_JSON, // 写入文件的数据是否 json 格式化
			Format:     "",                   // 如果写入文件的数据不 json 格式化，自定义日志格式
		}
		// 添加 file 为 logger 的一个输出
		logger.Attach("file", config.GAME_LOG_LEVEL, fileConfig)
	}
}

//llog emergency level
func Emergency(msg string) {
	logger.Emergency(msg)
	msg = "Emergencyf: " + msg
	go reportMail(msg)
}

//llog emergency format
func Emergencyf(format string, a ...interface{}) {
	logger.Emergencyf(format, a...)
	msg := fmt.Sprintf(format, a...)
	msg = "Emergencyf: " + msg
	go reportMail(msg)
}

//llog alert level
func Alert(msg string) {
	logger.Alert(msg)
	msg = "Alert: " + msg
	go reportMail(msg)
}

//llog alert format
func Alertf(format string, a ...interface{}) {
	logger.Alertf(format, a...)
	msg := fmt.Sprintf(format, a...)
	msg = "Alertf: " + msg
	go reportMail(msg)
}

//llog critical level
func Critical(msg string) {
	logger.Critical(msg)
	msg = "Critical: " + msg
	go reportMail(msg)
}

//llog critical format
func Criticalf(format string, a ...interface{}) {
	logger.Criticalf(format, a...)
	msg := fmt.Sprintf(format, a...)
	msg = "Criticalf: " + msg
	go reportMail(msg)
}

//llog error level
func Error(msg string) {
	logger.Error(msg)
	msg = "Error: " + msg
	go reportMail(msg)
}

//llog error format
func Errorf(format string, a ...interface{}) {
	logger.Errorf(format, a...)
	msg := fmt.Sprintf(format, a...)
	msg = "Error: " + msg
	go reportMail(msg)
}

//llog warning level
func Warning(msg string) {
	logger.Warning(msg)
}

//llog warning format
func Warningf(format string, a ...interface{}) {
	logger.Warningf(format, a...)
}

//llog notice level
func Notice(msg string) {
	logger.Notice(msg)
}

//llog notice format
func Noticef(format string, a ...interface{}) {
	logger.Noticef(format, a...)
}

//llog info level
func Info(msg string) {
	logger.Info(msg)
}

//llog info format
func Infof(format string, a ...interface{}) {
	logger.Infof(format, a...)
}

//llog debug level
func Debug(msg string) {
	logger.Debug(msg)
}

//llog debug format
func Debugf(format string, a ...interface{}) {
	logger.Debugf(format, a...)
}

//llog Fatal
func Fatal(msg string) {
	Error(msg)
	log.Fatal(msg)
}

//llog Fatal
func Fatalf(msg string, a ...interface{}) {
	Errorf(msg, a)
	log.Fatalf(msg, a)
}

func reportMail(errstr string) {
	reqParam := &struct {
		Tag     int    `json:"tag"`     //邮件类型
		Id      int    `json:"id"`      //区服id
		Content string `json:"content"` //邮件内容
	}{}
	reqParam.Tag = define.MAIL_TYPE_ERR
	reqParam.Id = config.NET_NODE_ID
	reqParam.Content = fmt.Sprintf("uid: %d \nname: %s\nhost: %s\r\n%s", config.SERVER_NODE_UID, config.SERVER_NAME, config.NET_GATE_SADDR, errstr)
	buffer, err := json.Marshal(&reqParam)
	if err == nil {
		lnats.Publish(define.TOPIC_SERVER_MAIL, buffer)
	}
}
