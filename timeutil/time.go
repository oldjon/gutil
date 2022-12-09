package timeutil

import (
	"time"
)

// GetTimeUnix using local time zone
func GetTimeUnix(t time.Time) int64 {
	return t.Local().Unix()
}

// GetNowTimeUnix using local time zone
func GetNowTimeUnix() int64 {
	return GetTimeUnix(time.Now())
}

// GetNowTimeStr string format "2006-01-02 15:04:05"
func GetNowTimeStr() string {
	return TimeToDateTime(time.Now())
}

//GetNowDateStr string format "2006-01-02"
func GetNowDateStr() string {
	return TimeToDate(time.Now())
}

// GetTimeFromUnix using local time zone
func GetTimeFromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// TodayXHourTime using local time zone
func TodayXHourTime(addHour time.Duration) time.Time {
	dateStr := TimeToDate(time.Now())
	t := DateStrToTime(dateStr)
	return t.Add(time.Hour * addHour)
}

// TodayXHourTimeUnix using local time zone
func TodayXHourTimeUnix(addHour time.Duration) int64 {
	dateStr := TimeToDate(time.Now())
	t := DateStrToTime(dateStr)
	t2 := t.Add(time.Hour * addHour)
	return GetTimeUnix(t2)
}

// TomorrowXHourTime using local time zone
func TomorrowXHourTime(addHour time.Duration) time.Time {
	dateStr := TimeToDate(time.Now())
	t := DateStrToTime(dateStr)
	return t.Add(time.Hour * (24 + addHour))
}

// TomorrowXHourTimeByLocation using local time zone
func TomorrowXHourTimeByLocation(addHour time.Duration, loc *time.Location) time.Time {
	dateStr := TimeToDate(time.Now().In(loc))
	t := DateStrToTimeByLocation(dateStr, loc)
	return t.Add(time.Hour * (24 + addHour))
}

// TodayXHourTimeByLocation using local time zone
func TodayXHourTimeByLocation(addHour time.Duration, loc *time.Location) time.Time {
	dateStr := TimeToDate(time.Now().In(loc))
	t := DateStrToTimeByLocation(dateStr, loc)
	return t.Add(time.Hour * addHour)
}

// TomorrowXHourTimeUnix using local time zone
func TomorrowXHourTimeUnix(addHour time.Duration) int64 {
	dateStr := TimeToDate(time.Now())
	t := DateStrToTime(dateStr)
	t2 := t.Add(time.Hour * (24 + addHour))
	return GetTimeUnix(t2)
}

// LastXHourTime using local time zone
func LastXHourTime(addHour time.Duration) time.Time {
	dateStr := TimeToDate(time.Now().Add(-time.Hour * 24))
	t := DateStrToTime(dateStr)
	return t.Add(time.Hour * addHour)
}

// LastXHourTimeUnix using local time zone
func LastXHourTimeUnix(addHour time.Duration) int64 {
	dateStr := TimeToDate(time.Now().Add(-time.Hour * 24))
	t := DateStrToTime(dateStr)
	t2 := t.Add(time.Hour * addHour)
	return GetTimeUnix(t2)
}

// ThisMondayXHourTimeUnix using local time zone
func ThisMondayXHourTimeUnix(addHour time.Duration) int64 {
	now := time.Now()
	weekDay := now.Weekday()
	if weekDay == time.Sunday {
		now = now.AddDate(0, 0, -6)
	} else {
		now = now.AddDate(0, 0, -int(weekDay)+1)
	}
	dateStr := TimeToDate(now)
	t := DateStrToTime(dateStr)
	t2 := t.Add(time.Hour * addHour)
	return GetTimeUnix(t2)
}

//AddDateToTime add year,month and date
func AddDateToTime(t int64, addYear int, addMonth int, addDate int) int64 {
	ti := GetTimeFromUnix(t)
	ti = ti.AddDate(addYear, addMonth, addDate)
	return GetTimeUnix(ti)
}

//AddSecondToNowTime add seconds to now time
func AddSecondToNowTime(addSecs int) string {
	now := time.Now()
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
func DateStrToTime(d string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", d, time.Local)
	return t
}

// DateStrToTimeByLocation string format "2006-01-02"
func DateStrToTimeByLocation(d string, loc *time.Location) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", d, loc)
	return t
}

// DateStrToTimeUnix string format "2006-01-02" to timestamp
func DateStrToTimeUnix(d string) int64 {
	if d == "" {
		return 0
	}

	t, _ := time.ParseInLocation("2006-01-02", d, time.Local)
	return t.Unix()
}

// DateTimeStrToTime string format "2006-01-02 15:04:05"
func DateTimeStrToTime(dt string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", dt, time.Local)
	return t
}

// DateTimeStrToTimeUnix string format "2006-01-02 15:04:05" to timestamp
func DateTimeStrToTimeUnix(dt string) int64 {
	if dt == "" {
		return 0
	}

	t, _ := time.ParseInLocation("2006-01-02 15:04:05", dt, time.Local)
	return t.Unix()
}

// DateTimeStrToTimeUnix string format "2006-01-02 15:04:05" to timestamp
func DateTimeStrToTimeUnixWithErr(dt string) (int64, error) {
	if dt == "" {
		return 0, nil
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05", dt, time.Local)
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

//LastRegionWeekRefreshTime return the last refresh time for the event's refresh weekly based on region
//For exmaple, if the refresh time is Monday 08:00 in each region and check region is -0500 EST
//If t is 2018-04-02 12:59 0000 UTC which is 2018-04-02 07:59 -0500 EST would return 2018-03-26 08:00 -0500 EST
//If t is 2018-04-02 13:01 0000 UTC which is 2018-04-02 08:01 -0500 EST would return 2018-04-02 08:00 -0500 EST
func LastRegionWeekRefreshTime(t time.Time, refreshDay time.Weekday, refreshHour int, refreshMinutes int, region *time.Location) time.Time {
	tt := t.In(region)
	weekday := tt.Weekday()
	year, month, day := tt.Date()
	hour, minute, _ := tt.Clock()
	if weekday < refreshDay ||
		(refreshDay == weekday && hour < refreshHour) ||
		(refreshDay == weekday && hour == refreshHour && minute <= refreshMinutes) {
		return time.Date(year, month, day+int(refreshDay-weekday)-7, refreshHour, refreshMinutes, 0, 0, region)
	}
	return time.Date(year, month, day-int(weekday-refreshDay), refreshHour, refreshMinutes, 0, 0, region)
}

// RegionMonday5HourDateTime return region Monday 05am date, format: 2006-01-02 04:00:00
func RegionMonday4HourDateTime(region *time.Location) string {
	now := time.Now().In(region)
	weekDay := now.Weekday()
	if weekDay == time.Sunday {
		now = now.AddDate(0, 0, -6)
	} else {
		now = now.AddDate(0, 0, -int(weekDay)+1)
	}
	dateStr := TimeToDate(now) + " 04:00:00"
	return dateStr
}
