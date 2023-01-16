package llog

import (
	"encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	go_logger "github.com/phachon/go-logger"
	"github.com/snowyyj001/loumiao/config"
	"github.com/snowyyj001/loumiao/define"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const ( //日志树输出级别
	DebugLevel = iota
	InfoLevel
	NoticeLevel
	WarnLevel
	ErrorLevel
)

var (
	clogger     *go_logger.Logger //zap的控制台颜色输出效果不好，这里用go_logger
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
	logLevel    int
	//remoteLog   *LogRemoteWrite

	ReportMailHanlder func (tag int, str string)
)

/*
type LogRemoteWrite struct {
	clientLog  net.Conn
	lineNum    int64
	remoteAddr string
}

func (l *LogRemoteWrite) connectUdp() error {
	if l.clientLog != nil {
		l.clientLog.Close()
		l.clientLog = nil
	}
	l.remoteAddr = config.NET_LOG_SADDR()
	if len(l.remoteAddr) == 0 {
		return fmt.Errorf("no log udp server")
	}
	if conn, err := net.Dial("udp", l.remoteAddr); err == nil {
		l.clientLog = conn
	} else {
		return err
	}
	return nil
}

func (l *LogRemoteWrite) Write(p []byte) (n int, err error) {
	if l.clientLog == nil {
		err = l.connectUdp()
		if err != nil {
			err = fmt.Errorf(string(p[:]))
			return
		}
	}
	bstream := base.NewBitStream_1(len(p) + len(config.SERVER_NAME) + 64)
	var sb strings.Builder
	atomic.AddInt64(&l.lineNum, 1)
	sb.WriteString(strconv.Itoa(int(l.lineNum)))
	sb.WriteString(" ")
	sb.Write(p)
	bstream.WriteString(config.SERVER_NAME)
	bstream.WriteString(sb.String())
	n, err = l.clientLog.Write(bstream.GetBuffer())
	if err != nil {
		//udp的connect并不真和远端server建立连接，只是在内核保存了对端的ip和端口
		//检查是否存在立即可知的错误（一个明显不可达的目的地）
		//接收端应该设置SO_RCVBUF，较大的接收缓冲区可以适当减少丢包，但是一般不需要设置，因为大部分linux系统默认值就是可设置的最大值
		//发送端无需设置SO_SNDBUF，因为udp whatever you send goes directly out to the network
		if l.connectUdp() == nil {
			n, err = l.clientLog.Write(bstream.GetBuffer())
			if err != nil {
				fmt.Println("0.write filecollect log failed: ",err.Error())
			}
		} else {
			fmt.Println("1.write filecollect log failed: ",err.Error())
		}
	}
	return
}
*/
func getLogWriter(fileName string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fileName, //日志文件的位置
		MaxSize:    100,      //日志文件的最大大小（以MB为单位）
		MaxBackups: 100,      //保留旧文件的最大个数
		MaxAge:     7,        //保留旧文件的最大天数
		Compress:   false,    //是否压缩/归档旧文件
	}

	return zapcore.AddSync(lumberJackLogger)
	/*if config.GAME_LOG_CONLOSE {
		//return zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(os.Stdout))
		return zapcore.AddSync(lumberJackLogger)
	} else {
		remoteLog = new(LogRemoteWrite)
		remoteLog.connectUdp()
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(remoteLog))
		//return zapcore.AddSync(lumberJackLogger)
	}*/
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func init() {
	//os.Mkdir("logs", os.ModePerm)
	//os.Mkdir(fmt.Sprintf("logs/%s", config.SERVER_NAME), os.ModePerm)
	SetLevel(config.GAME_LOG_LEVEL)
	filename := fmt.Sprintf("./logs/%s/%s.log", config.SERVER_TYPE_NAME, config.SERVER_NAME)
	core := zapcore.NewCore(getEncoder(), getLogWriter(filename), zapcore.DebugLevel)
	if config.GAME_LOG_CONLOSE {
		logger = zap.New(core)

		clogger = go_logger.NewLogger()
		// 命令行输出配置
		consoleConfig := &go_logger.ConsoleConfig{
			Color:      true,  // 命令行输出字符串是否显示颜色
			JsonFormat: false, // 命令行输出字符串是否格式化
			Format:     "",    // 如果输出的不是 json 字符串，JsonFormat: false, 自定义输出的格式
		}
		clogger.Detach("console")
		clogger.Attach("console", 7, consoleConfig)
	} else {
		logger = zap.New(core)
	}
	sugarLogger = logger.Sugar()
}

//llog Panic level
func Panic(msg string) {
	ReportMailHanlder(define.MAIL_TYPE_ERR, msg)
	logger.Panic(msg)
}

//llog Panic format
func Panicf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Panic(msg)
}

//llog error level
func Error(msg string) {
	go ReportMailHanlder(define.MAIL_TYPE_ERR, msg)
	logger.Error(msg)
	if clogger != nil {
		clogger.Error(msg)
	}
}

//llog error format
func Errorf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Error(msg)
}

//llog notice level
func Notice(msg string) {
	go ReportMailHanlder(define.MAIL_TYPE_WARING, msg)
	logger.Warn(msg)
	if clogger != nil {
		clogger.Notice(msg)
	}
}

//llog Notice format
func Noticef(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Notice(msg)
}

//llog warning level
func Warning(msg string) {
	if logLevel >= ErrorLevel {
		return
	}
	logger.Warn(msg)
	if clogger != nil {
		clogger.Warning(msg)
	}
}

//llog warning format
func Warningf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Warning(msg)
}

//llog info level
func Info(msg string) {
	if logLevel >= WarnLevel {
		return
	}
	logger.Info(msg)
	if clogger != nil {
		clogger.Info(msg)
	}
}

//llog info format
func Infof(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Info(msg)
}

//llog debug level
func Debug(msg string) {
	if logLevel >= InfoLevel {
		return
	}
	logger.Debug(msg)
	if clogger != nil {
		clogger.Debug(msg)
	}
}

//llog debug format
func Debugf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Debug(msg)
}

//llog Fatal
func Fatal(msg string) {
	ReportMailHanlder(define.MAIL_TYPE_ERR, msg)
	logger.Fatal(msg)
}

//llog Fatal
func Fatalf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	Fatal(msg)
}

func SetLevel(_level int) {
	logLevel = _level
}

func Tp_SetLevel(data []byte) {
	reqParam := &struct {
		Uid   int `json:"uid"`
		Level int `json:"level"`
	}{}
	if err := json.Unmarshal(data, reqParam); err == nil {
		if reqParam.Uid == config.SERVER_NODE_UID {
			Infof("Tp_SetLevel: oldlevel=%d, newlevel=%d", logLevel, reqParam.Level)
			SetLevel(reqParam.Level)
		}
	}
}