package log

import (
	"github.com/snowyyj001/loumiao/config"

	"github.com/phachon/go-logger"
)

var (
	logger *go_logger.Logger
)

func init() {
	logger = go_logger.NewLogger()
	logger.Detach("console")

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
		// 文件输出配置
		fileConfig := &go_logger.FileConfig{
			Filename:   "./logs/server.log",  // 日志输出文件名，不自动存在
			MaxSize:    0,                    // 文件最大值（KB），默认值0不限
			MaxLine:    100000,               // 文件最大行数，默认 0 不限制
			DateSlice:  "d",                  // 文件根据日期切分， 支持 "Y" (年), "m" (月), "d" (日), "H" (时), 默认 "no"， 不切分
			JsonFormat: config.GAME_LOG_JSON, // 写入文件的数据是否 json 格式化
			Format:     "",                   // 如果写入文件的数据不 json 格式化，自定义日志格式
		}
		// 添加 file 为 logger 的一个输出
		logger.Attach("file", config.GAME_LOG_LEVEL, fileConfig)
	}
}

//log emergency level
func Emergency(msg string) {
	logger.Emergency(msg)
}

//log emergency format
func Emergencyf(format string, a ...interface{}) {
	logger.Emergencyf(format, a...)
}

//log alert level
func Alert(msg string) {
	logger.Alert(msg)
}

//log alert format
func Alertf(format string, a ...interface{}) {
	logger.Alertf(format, a...)
}

//log critical level
func Critical(msg string) {
	logger.Critical(msg)
}

//log critical format
func Criticalf(format string, a ...interface{}) {
	logger.Criticalf(format, a...)
}

//log error level
func Error(msg string) {
	logger.Error(msg)
}

//log error format
func Errorf(format string, a ...interface{}) {
	logger.Errorf(format, a...)
}

//log warning level
func Warning(msg string) {
	logger.Warning(msg)
}

//log warning format
func Warningf(format string, a ...interface{}) {
	logger.Warningf(format, a...)
}

//log notice level
func Notice(msg string) {
	logger.Notice(msg)
}

//log notice format
func Noticef(format string, a ...interface{}) {
	logger.Noticef(format, a...)
}

//log info level
func Info(msg string) {
	logger.Info(msg)
}

//log info format
func Infof(format string, a ...interface{}) {
	logger.Infof(format, a...)
}

//log debug level
func Debug(msg string) {
	logger.Debug(msg)
}

//log debug format
func Debugf(format string, a ...interface{}) {
	logger.Debugf(format, a...)
}
