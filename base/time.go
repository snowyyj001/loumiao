package base

import (
	"time"
)

var dayOffset int32
var hourOffset int32
var minuteOffset int32

var TimeOffset int
var TimeOffsetD time.Duration

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
