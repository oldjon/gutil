package timeutil

import (
	"time"
)

// GetNowTimeStr string format "2006-01-02 15:04:05" if loc is nil, using local time zone
func GetNowTimeStr(loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	return TimeToDateTime(time.Now().In(loc))
}

// GetNowDateStr string format "2006-01-02" if loc is nil, using local time zone
func GetNowDateStr(loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	return TimeToDate(time.Now().In(loc))
}

// GetTimeFromUnix
func GetTimeFromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// TodayXHourTime if loc is nil, using local time zone
func TodayXHourTime(addHour time.Duration, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	dateStr := TimeToDate(time.Now().In(loc))
	return DateStrToTime(dateStr, loc).Add(time.Hour * addHour)
}

// TodayXHourTimeUnix if loc is nil, using local time zone
func TodayXHourTimeUnix(addHour time.Duration, loc *time.Location) int64 {
	if loc == nil {
		loc = time.Local
	}
	dateStr := TimeToDate(time.Now().In(loc))
	return DateStrToTime(dateStr, loc).Add(time.Hour * addHour).Unix()
}

// TomorrowXHourTime if loc is nil, using local time zone
func TomorrowXHourTime(addHour time.Duration, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	dateStr := TimeToDate(time.Now().In(loc))
	return DateStrToTime(dateStr, loc).Add(time.Hour * (24 + addHour))
}

// TomorrowXHourTimeUnix if loc is nil, using local time zone
func TomorrowXHourTimeUnix(addHour time.Duration, loc *time.Location) int64 {
	if loc == nil {
		loc = time.Local
	}
	dateStr := TimeToDate(time.Now().In(loc))
	return DateStrToTime(dateStr, loc).Add(time.Hour * (24 + addHour)).Unix()
}

// LastXHourTime if loc is nil, using local time zone
func LastXHourTime(addHour time.Duration, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	dateStr := TimeToDate(time.Now().In(loc).Add(-time.Hour * 24))
	return DateStrToTime(dateStr, loc).Add(time.Hour * addHour)
}

// LastXHourTimeUnix if loc is nil, using local time zone
func LastXHourTimeUnix(addHour time.Duration, loc *time.Location) int64 {
	if loc == nil {
		loc = time.Local
	}
	dateStr := TimeToDate(time.Now().In(loc).Add(-time.Hour * 24))
	return DateStrToTime(dateStr, loc).Add(time.Hour * addHour).Unix()
}

// ThisMondayXHourTimeUnix if loc is nil, using local time zone
func ThisMondayXHourTimeUnix(addHour time.Duration, loc *time.Location) int64 {
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	weekDay := now.Weekday()
	if weekDay == time.Sunday {
		now = now.AddDate(0, 0, -6)
	} else {
		now = now.AddDate(0, 0, -int(weekDay)+1)
	}
	dateStr := TimeToDate(now)
	return DateStrToTime(dateStr, loc).Add(time.Hour * addHour).Unix()
}

// AddDateToTime add year,month and date, if loc is nil, using local time zone
func AddDateToTime(t int64, addYear int, addMonth int, addDate int, loc *time.Location) int64 {
	if loc == nil {
		loc = time.Local
	}
	ti := GetTimeFromUnix(t).In(loc)
	ti = ti.AddDate(addYear, addMonth, addDate)
	return ti.Unix()
}

// AddSecondToNowTime add seconds to now time, if loc is nil, using local time zone
func AddSecondToNowTime(addSecs int, loc *time.Location) string {
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	newTime := now.Add(time.Second * time.Duration(addSecs))
	return TimeToDateTime(newTime)
}

// TimeToDate format "2006-01-02"
func TimeToDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// TimeToDateTime format "2006-01-02 15:04:05"
func TimeToDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// DateStrToTime string format "2006-01-02"
func DateStrToTime(d string, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	t, _ := time.ParseInLocation("2006-01-02", d, loc)
	return t
}

// DateStrToTimeUnix string format "2006-01-02" to timestamp
func DateStrToTimeUnix(d string, loc *time.Location) int64 {
	if d == "" {
		return 0
	}
	if loc == nil {
		loc = time.Local
	}
	t, _ := time.ParseInLocation("2006-01-02", d, loc)
	return t.Unix()
}

// DateTimeStrToTime string format "2006-01-02 15:04:05"
func DateTimeStrToTime(dt string, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.Local
	}
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", dt, loc)
	return t
}

// DateTimeStrToTimeUnix string format "2006-01-02 15:04:05" to timestamp
func DateTimeStrToTimeUnix(dt string, loc *time.Location) int64 {
	if dt == "" {
		return 0
	}
	if loc == nil {
		loc = time.Local
	}
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", dt, loc)
	return t.Unix()
}

// DateTimeStrToTimeUnix string format "2006-01-02 15:04:05" to timestamp
func DateTimeStrToTimeUnixWithErr(dt string, loc *time.Location) (int64, error) {
	if dt == "" {
		return 0, nil
	}
	if loc == nil {
		loc = time.Local
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", dt, loc)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// IsWeekend check time sunday or saturday
func IsWeekend(t time.Time) bool {
	wd := t.Weekday()
	if wd == time.Sunday || wd == time.Saturday {
		return true
	}
	return false
}

// LastRegionWeekRefreshTime return the last refresh time for the event's refresh weekly based on region
// For exmaple, if the refresh time is Monday 08:00 in each region and check region is -0500 EST
// If t is 2018-04-02 12:59 0000 UTC which is 2018-04-02 07:59 -0500 EST would return 2018-03-26 08:00 -0500 EST
// If t is 2018-04-02 13:01 0000 UTC which is 2018-04-02 08:01 -0500 EST would return 2018-04-02 08:00 -0500 EST
func LastRegionWeekRefreshTime(t time.Time, refreshDay time.Weekday, refreshHour int, refreshMinutes int, loc *time.Location) time.Time {
	tt := t.In(loc)
	weekday := tt.Weekday()
	year, month, day := tt.Date()
	hour, minute, _ := tt.Clock()
	if weekday < refreshDay ||
		(refreshDay == weekday && hour < refreshHour) ||
		(refreshDay == weekday && hour == refreshHour && minute <= refreshMinutes) {
		return time.Date(year, month, day+int(refreshDay-weekday)-7, refreshHour, refreshMinutes, 0, 0, loc)
	}
	return time.Date(year, month, day-int(weekday-refreshDay), refreshHour, refreshMinutes, 0, 0, loc)
}

// RegionMonday5HourDateTime return region Monday 05am date, format: 2006-01-02 04:00:00
func RegionMonday4HourDateTime(loc *time.Location) string {
	now := time.Now().In(loc)
	weekDay := now.Weekday()
	if weekDay == time.Sunday {
		now = now.AddDate(0, 0, -6)
	} else {
		now = now.AddDate(0, 0, -int(weekDay)+1)
	}
	dateStr := TimeToDate(now) + " 04:00:00"
	return dateStr
}
