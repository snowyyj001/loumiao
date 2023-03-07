package base

import (
	"time"
)

var (
	timeOffset    time.Duration
	StartTick     int64
	NowTick       int64
	Timestamp     int64 // 当前秒数
	TIME_LOCATION *time.Location
)

const (
	DAY_SECONDS  int64  = 86400          //每日秒数
	SECOND_MILLI int64  = 1000           //1秒钟,毫秒
	MINITE_MILLI int64  = 1000 * 60      //1分钟,毫秒
	HOUR_MILLI   int64  = 1000 * 60 * 60 //1小时,毫秒
	RFC3339Day   string = "2006_01_02"   //RFC3339格式，精确到天
)

func init() {
	timerTick()
}

func timerTick() {
	now := TimeNow()
	StartTick = now.UnixNano() / 1000000
	NowTick = StartTick
	Timestamp = NowTick / 1000
	lastTimestamp := Timestamp
	var ticker = time.NewTicker(time.Millisecond)
	go func() {
		for true {
			select {
			case <-ticker.C:
				now = TimeNow()
				NowTick = now.UnixNano() / 1000000
				Timestamp = NowTick / 1000
				if Timestamp != lastTimestamp {
					lastTimestamp = Timestamp

				}
			}
		}
		ticker.Stop()
	}()
}

func TimeNow() time.Time {
	if timeOffset > 0 {
		nowTime := time.Now()
		nowTime = nowTime.Add(timeOffset)
		return nowTime
	}
	return time.Now()
}

func TimeForDelete() time.Time {
	return TimeNow().Truncate(time.Hour * 24)
}

func SetOffsetTime(offset time.Duration) {
	timeOffset = offset
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
	t := TimeNow()
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
	t := TimeNow()
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

// GetDayTime 获取当天的0点和24点时间
// @st：指定那一天，0默认当天
func GetDayTime(st int64) (int64, int64) {
	var timeStr string
	if st == 0 {
		timeStr = TimeNow().Format(RFC3339Day)
	} else {
		timeStr = time.Unix(st, 0).Format(RFC3339Day)
	}
	t, _ := time.ParseInLocation(RFC3339Day, timeStr, TIME_LOCATION)
	var beginTimeNum = t.Unix()
	var endTimeNum = beginTimeNum + DAY_SECONDS
	return beginTimeNum, endTimeNum
}

// IsSameDay 是否是同一天
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

// TimeStr 当前格式化时间字符串
func TimeStr() string {
	return TimeNow().Format("2006-01-02 15:04:05")
}

// TimeStrFormat 当前格式化时间字符串
func TimeStrFormat(mat string) string {
	return TimeNow().Format(mat)
}

// TimeStampSec 时间戳秒
func TimeStampSec() int64 {
	return Timestamp / 1000
}

// TimeStamp 时间戳毫秒
func TimeStamp() int64 {
	return Timestamp
}

// TimeStampTarget 指定日期的时间戳毫秒
func TimeStampTarget(y int, m time.Month, d int, h int, mt int, s int) int64 {
	return time.Date(y, m, d, h, mt, s, 0, time.Local).UnixNano() / int64(time.Millisecond)
}

// DifferDaysOnDate 计算两个时间戳在日期上相差的天数
// now, old 时间戳,秒
// offset 零点偏移时间
// return days 在日期上相差的天数
func DifferDaysOnDate(now, old int64, offset int) int {
	start := time.Unix(now, 0).Add(-time.Duration(offset) * time.Hour)
	end := time.Unix(old, 0).Add(-time.Duration(offset) * time.Hour)

	day1 := start.YearDay()
	day2 := end.YearDay()
	return day1 - day2
}
