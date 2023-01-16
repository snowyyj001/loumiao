package base

import (
	"time"
)

var dayOffset int32
var hourOffset int32
var minuteOffset int32

var TimeOffset int
var TimeOffsetD time.Duration
var TIME_LOCATION *time.Location //时区

const (
	DAY_SECONDS  int64  = 86400        //每日秒数
	SECOND_MILLI int64  = 1000         //1秒钟,毫秒
	MINITE_MILLI int64  = 1000 * 60    //1分钟,毫秒
	RFC3339Day   string = "2006_01_02" //RFC3339格式，精确到天
)

func init() {
	_, TimeOffset = time.Now().Zone()
	TimeOffsetD = time.Duration(TimeOffset) * time.Second
}

func TimeNow() time.Time {
	if dayOffset != 0 || hourOffset != 0 || minuteOffset != 0 {
		return time.Now().Add(time.Duration(dayOffset)*time.Hour*24 + time.Duration(hourOffset)*time.Hour + time.Duration(minuteOffset)*time.Minute)
	}
	return time.Now()
}

func TimeNowUnix() uint32 {
	return uint32(TimeNow().Unix())
}

func TimeForDelete() time.Time {
	return TimeNow().Truncate(time.Hour * 24)
}

func SetTimeOffset(day int32, hour int32, minute int32) {
	dayOffset = day
	hourOffset = hour
	minuteOffset = minute
}

// 取当天0点
func Time0(timeIn time.Time) time.Time {
	return timeIn.Truncate(time.Hour * 24).Add(-TimeOffsetD)
}

func TimeToday0() time.Time {
	return Time0(TimeNow())
}

func TimeToday0Unix() uint32 {
	return uint32(TimeToday0().Unix())
}

func TimeMonday0(timeIn time.Time) time.Time {
	return timeIn.Add(TimeOffsetD).Truncate(time.Hour * 24 * 7).Add(-TimeOffsetD)
}

func TimeThisMonday0() time.Time {
	return TimeMonday0(TimeNow())
}

func WeekDayNormal() int {
	day := int(TimeNow().Weekday()) - 1
	if day < 0 {
		day = 6
	}
	return day
}

// IsToday 是否是当天
// @stamp 时间戳秒
func IsToday(stamp int64) bool {
	t := time.Now()
	_, timeOffset := t.Zone()
	timeD := time.Duration(timeOffset) * time.Second
	checkTimeStart := t.Add(timeD).Truncate(time.Hour * 24).Add(-timeD)
	checkTimeEnd := t.Add(timeD).Truncate(time.Hour * 24).Add(-timeD).Add(time.Hour * 24)
	if stamp >= checkTimeStart.Unix() && stamp <= checkTimeEnd.Unix() {
		return true
	}
	return false
}

// IsThisWeek stamp是否是本周
func IsThisWeek(stamp int64) bool {
	t := time.Now()
	_, timeOffset := t.Zone()
	timeD := time.Duration(timeOffset) * time.Second
	checkTimeStart := t.Add(timeD).Truncate(time.Hour * 24 * 7).Add(-timeD)
	checkTimeEnd := t.Add(timeD).Truncate(time.Hour * 24 * 7).Add(-timeD).Add(time.Hour * 24 * 7)
	if stamp >= checkTimeStart.Unix() && stamp <= checkTimeEnd.Unix() {
		return true
	}
	return false
}

// IsThisMonth stamp是否是本月
func IsThisMonth(stamp int64) bool {
	t := time.Now()
	currentYear, currentMonth, _ := t.Date()
	checkTimeStart := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, t.Location())
	checkTimeEnd := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0)
	if stamp >= checkTimeStart.Unix() && stamp <= checkTimeEnd.Unix() {
		return true
	}
	return false
}

// 获取当天的0点和24点时间
// @st：指定那一天，0默认当天
func GetDayTime(st int64) (int64, int64) {
	var timeStr string
	if st == 0 {
		timeStr = time.Now().Format(RFC3339Day)
	} else {
		timeStr = time.Unix(st, 0).Format(RFC3339Day)
	}
	t, _ := time.ParseInLocation(RFC3339Day, timeStr, TIME_LOCATION)
	var beginTimeNum = t.Unix()
	var endTimeNum = beginTimeNum + DAY_SECONDS
	return beginTimeNum, endTimeNum
}

// 是否是同一天
// 请确保stmp2 > stmp1
func IsSameDay(stmp1, stmp2 int64) bool {
	if stmp2-stmp1 >= DAY_SECONDS {
		return false
	}
	_, end := GetDayTime(stmp1)
	if stmp2 <= end {
		return false
	}
	return true
}

// 当前格式化时间字符串
func TimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 当前格式化时间字符串
func TimeStrFormat(mat string) string {
	return time.Now().Format(mat)
}

// 时间戳秒
func TimeStampSec() int64 {
	return time.Now().Unix()
}

// 时间戳毫秒
func TimeStamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// 指定日期的时间戳毫秒
func TimeStampTarget(y int, m time.Month, d int, h int, mt int, s int) int64 {
	return time.Date(y, m, d, h, mt, s, 0, TIME_LOCATION).UnixNano() / int64(time.Millisecond)
}
