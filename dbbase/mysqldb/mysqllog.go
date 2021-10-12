package mysqldb

import (
	"context"
	"fmt"
	"time"

	"github.com/snowyyj001/loumiao/llog"

	"gorm.io/gorm/logger"
)

type SqlLog struct {
	LogLevel logger.LogLevel
}

func newloger() *SqlLog {
	return &SqlLog{}
}

func (self *SqlLog) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *self
	newlogger.LogLevel = level
	return &newlogger
}

func (self *SqlLog) Info(ctx context.Context, msg string, data ...interface{}) {
	if self.LogLevel >= logger.Info {
		llog.Debugf("mysql : %s", fmt.Sprintf("msg=%s, data=%v", msg, data))
	}
}

func (self *SqlLog) Warn(ctx context.Context, msg string, data ...interface{}) {
	if self.LogLevel >= logger.Warn {
		llog.Warningf("mysql Warn: %s", fmt.Sprintf("msg=%s, data=%v", msg, data))
	}
}

func (self *SqlLog) Error(ctx context.Context, msg string, data ...interface{}) {
	if self.LogLevel >= logger.Error {
		llog.Errorf("mysql Error : %s", fmt.Sprintf("msg=%s, data=%v", msg, data))
	}
}

func (self *SqlLog) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if self.LogLevel <= logger.Silent {
		return
	}
	sql, rows := fc()
	llog.Debugf("musql trace: %s", fmt.Sprintf("sql=%s, rows=%d", sql, rows))
}
