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

//取当天0点
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
