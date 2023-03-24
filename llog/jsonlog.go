package llog

import (
	"encoding/json"
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/snowyyj001/loumiao/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

var (
	logger_json *zap.Logger
)

type LogEvent struct {
	Env      string                 `json:"env"`      //环境, 示例：prod/test
	Platform string                 `json:"platform"` //平台, 示例：android/ios
	Game     string                 `json:"game"`     //游戏代号，示例：rxjx
	Channel  string                 `json:"channel"`  //渠道, 示例：2144
	Version  string                 `json:"version"`  //游戏版本号, 示例:v1.0.0
	Ip       string                 `json:"ip"`
	Time     int64                  `json:"time"`
	Account  string                 `json:"account"`
	ServerId int32                  `json:"server_id,string"`
	RoleId   int32                  `json:"role_id,string"`
	Event    string                 `json:"event"`
	Data     map[string]interface{} `json:"data"`
}

type UserLog struct {
	Logs LogEvent
}

// AppendLog 附加日志信息
func (r *UserLog) AppendLog(platform string, channel string, ip string, account string, serverid int32) {
	r.Logs.Platform = platform
	r.Logs.Channel = channel
	r.Logs.Ip = ip
	r.Logs.Account = account
	r.Logs.ServerId = serverid
}

// 打印关卡测试
func (r *UserLog) printLog(event string) {

	r.Logs.Time = time.Now().Unix()
	r.Logs.Event = event

	if str, err := json.Marshal(&r.Logs); err == nil {
		logger_json.Info(string(str))
	}

	r.Logs.Data = make(map[string]interface{})
}

func (r *UserLog) LogInfo(event string, datas ...interface{}) bool {
	if len(datas)&1 != 0 {
		return false
	}

	for i := 0; i < len(datas); i = i + 2 {
		r.Logs.Data[datas[i].(string)] = datas[i+1]
	}

	r.printLog(event)
	return true
}

func getJsonLogWriter(fileName string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fileName, //日志文件的位置
		MaxSize:    1024,     //日志文件的最大大小（以MB为单位）
		MaxBackups: 100,      //保留旧文件的最大个数
		MaxAge:     7,        //保留旧文件的最大天数
		Compress:   false,    //是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberJackLogger)
}

func LogInfo(msg string) {
	logger_json.Info(msg)
}

func init() {
	//os.Mkdir("logs", os.ModePerm)
	//os.Mkdir(fmt.Sprintf("logs/%s", config.SERVER_NAME), os.ModePerm)
	SetLevel(config.GAME_LOG_LEVEL)
	filename := fmt.Sprintf("../jsonlogs/%s.log", config.SERVER_NAME)
	core := zapcore.NewCore(getEncoder(), getJsonLogWriter(filename), zapcore.DebugLevel)
	logger_json = zap.New(core)
}
