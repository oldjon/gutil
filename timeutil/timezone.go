package timeutil

import (
	"sync"
	"time"
)

const (
	secsPerHour = 3600
	secsPerDay  = 24 * 60 * 60
)

type timezoneConfig struct {
	defaultRegion         string
	regionLocationMapping map[string]*time.Location
}

var (
	globalTZConfig = &timezoneConfig{
		defaultRegion: "SG",
		regionLocationMapping: map[string]*time.Location{
			"VN":     time.FixedZone("VN", 7*secsPerHour),
			"TH":     time.FixedZone("TH", 7*secsPerHour),
			"ID":     time.FixedZone("ID", 7*secsPerHour),
			"TW":     time.FixedZone("TW", 8*secsPerHour),
			"BR":     time.FixedZone("BR", -3*secsPerHour),
			"SG":     time.FixedZone("SG", 8*secsPerHour),
			"US":     time.FixedZone("US", -5*secsPerHour),
			"RU":     time.FixedZone("RU", 3*secsPerHour),
			"EUROPE": time.FixedZone("EUROPE", 2*secsPerHour),
			"SAC":    time.FixedZone("SAC", -3*secsPerHour),
			"IND":    time.FixedZone("IND", 5.5*secsPerHour),
			"ME":     time.FixedZone("ME", 2*secsPerHour),
			"ZA":     time.FixedZone("ZA", 2*secsPerHour),
			"NA":     time.FixedZone("NA", -4*secsPerHour),
			"PK":     time.FixedZone("PK", 5*secsPerHour),
			"BD":     time.FixedZone("BD", 6*secsPerHour),
			"AR":     time.FixedZone("AR", -3*secsPerHour),
			"JP":     time.FixedZone("JP", 9*secsPerHour),
		},
	}

	globalMux sync.RWMutex
)

func getGlobalTc() *timezoneConfig {
	globalMux.RLock()
	defer globalMux.RUnlock()
	return globalTZConfig
}

func setGlobalTc(tc *timezoneConfig) {
	globalMux.Lock()
	globalTZConfig = tc
	globalMux.Unlock()
}

func (tc *timezoneConfig) getDefaultRegion() string {
	return tc.defaultRegion
}

func (tc *timezoneConfig) regionTimeZoneLocation(region string) (*time.Location, error) {
	if loc, ok := tc.regionLocationMapping[region]; ok {
		return loc, nil
	}
	return tc.regionLocationMapping[tc.defaultRegion], nil
}

func (tc *timezoneConfig) setDefaultRegion(region string) error {
	if _, ok := tc.regionLocationMapping[region]; !ok {
		return errDefaultNotInMapping
	}
	tc.defaultRegion = region
	return nil
}

func (tc *timezoneConfig) clone() *timezoneConfig {
	newTc := &timezoneConfig{
		defaultRegion:         tc.defaultRegion,
		regionLocationMapping: make(map[string]*time.Location, len(tc.regionLocationMapping)),
	}
	for region, loc := range tc.regionLocationMapping {
		newTc.regionLocationMapping[region] = loc
	}
	return newTc
}

// PatchDefaultRegion change the default region for the pkg
func PatchDefaultRegion(region string) error {
	newTc := getGlobalTc().clone()
	err := newTc.setDefaultRegion(region)
	if err != nil {
		return err //nolint
	}

	setGlobalTc(newTc)
	return nil
}

// PatchRegionLocationMapping change the default region location mapping for the pkg
func PatchRegionLocationMapping(options ...*PatchOption) {
	newTc := getGlobalTc().clone()
	for _, option := range options {
		newTc.regionLocationMapping[option.Region] = time.FixedZone(option.Region, int(option.OffsetHour*secsPerHour))
	}

	setGlobalTc(newTc)
}

type PatchOption struct {
	Region     string
	OffsetHour float32
}

func GetDefaultRegion() string {
	return getGlobalTc().getDefaultRegion()
}

// RegionTimeZoneLocation : return time zone location
func RegionTimeZoneLocation(region string) (*time.Location, error) {

	return getGlobalTc().regionTimeZoneLocation(region)
}

// LocalDateTime : convert GMT+8 datetime to Local datetime, format yyyy-mm-dd hh:ii:ss
func LocalDateTime(datetime string, region string) (string, error) {
	return convLocalDateTime(datetime, region, "2006-01-02 15:04:05")
}

// LocalDate : convert GMT+8 datetime to Local datetime, format yyyy-mm-dd
func LocalDate(datetime string, region string) (string, error) {
	return convLocalDateTime(datetime, region, "2006-01-02")
}

// LocalNowDateTime : get local datetime now, format yyyy-mm-dd hh:ii:ss
func LocalNowDateTime(region string) (string, error) {
	return convLocalNowDateTime(region, "2006-01-02 15:04:05")
}

// LocalNowDate : get local datetime now, format yyyy-mm-dd
func LocalNowDate(region string) (string, error) {
	return convLocalNowDateTime(region, "2006-01-02")
}

// LocalDateTimeFromTimestamp : convert timestamp to Local datetime, format yyyy-mm-dd hh:ii:ss
func LocalDateTimeFromTimestamp(timestamp int64, region string) (string, error) {
	return convLocalDateTimeFromTimeStamp(timestamp, region, "2006-01-02 15:04:05")
}

// LocalDateFromTimestamp : convert timestamp to Local datetime, format yyyy-mm-dd
func LocalDateFromTimestamp(timestamp int64, region string) (string, error) {
	return convLocalDateTimeFromTimeStamp(timestamp, region, "2006-01-02")
}

// LocalMonthFromTimestamp : convert timestamp to Local datetime, format yyyy-mm
func LocalMonthFromTimestamp(timestamp int64, region string) (string, error) {
	return convLocalDateTimeFromTimeStamp(timestamp, region, "2006-01")
}

// LocalDateTimeToTimeStamp : convert datetime to timestamp, format yyyy-mm-dd hh:ii:ss
func LocalDateTimeToTimeStamp(datetime string, region string) (int64, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0, err
	}
	dt, err := time.ParseInLocation("2006-01-02 15:04:05", datetime, loc)
	if err != nil {
		return 0, err
	}
	return dt.Unix(), nil
}

// LocalDateToTimeStamp : convert datetime to timestamp, format yyyy-mm-dd
func LocalDateToTimeStamp(datetime string, region string) (int64, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0, err
	}
	dt, err := time.ParseInLocation("2006-01-02", datetime, loc)
	if err != nil {
		return 0, err
	}
	return dt.Unix(), nil
}

// LocalDateTimeToLocalDate : convert datetime to date, from yyyy-mm-dd hh:ii:ss to yyyy-mm-dd
func LocalDateTimeToLocalDate(datetimeStr string, offsetHour int64, region string) (string, error) {
	var dateStr string
	timestamp, err := LocalDateTimeToTimeStamp(datetimeStr, region)
	if err != nil {
		return dateStr, err
	}

	dt := time.Unix(timestamp, 0)
	if dt.Hour() >= int(offsetHour) {
		dateStr, err = LocalDateFromTimestamp(timestamp, region)
		if err != nil {
			return dateStr, err
		}
		return dateStr, nil
	}

	// last day
	dateStr, err = LocalDateFromTimestamp(timestamp-offsetHour*secsPerHour, region)
	if err != nil {
		return dateStr, err
	}
	return dateStr, nil
}

// LocalDateByAddSomeDays : 2016-01-02 + (2days-1) = 2016-01-03
func LocalDateByAddSomeDays(dateStr string, someDays int64, region string) (string, error) {
	var currentDate string
	timestamp, err := LocalDateToTimeStamp(dateStr, region)
	if err != nil {
		return currentDate, err
	}
	if someDays > 0 {
		someDays = someDays - 1
	}
	currentDate, err = LocalDateFromTimestamp(timestamp+someDays*secsPerDay, region)
	if err != nil {
		return currentDate, err
	}

	return currentDate, nil
}

// DayLengthFromTwoLocalDate :
func DayLengthFromTwoLocalDate(dateStr1 string, dateStr2 string, region string) (uint32, error) {
	var dayLength uint32
	timestamp1, err := LocalDateToTimeStamp(dateStr1, region)
	if err != nil {
		return dayLength, err
	}
	timestamp2, err := LocalDateToTimeStamp(dateStr2, region)
	if err != nil {
		return dayLength, err
	}
	dayLength = uint32((timestamp2-timestamp1)/secsPerDay) + 1
	return dayLength, nil
}

// GetOffsetToUTCInSecs : calculate the offset in seconds between the input region and UTC
func GetOffsetToUTCInSecs(region string) (int64, error) {
	now := time.Now()
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0, err
	}
	_, offset := now.In(loc).Zone()
	return int64(offset), nil
}

// LocalWeekday : get local weekday. Monday = 1, Sunday = 7
func LocalWeekday(region string) (int32, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0, err
	}
	weekDay := time.Now().In(loc).Weekday()
	if weekDay == time.Sunday {
		return 7, nil
	}
	return int32(weekDay), nil
}

// LocalWeekdayFromTimestamp : get local weekday by timestamp. Monday = 1, Sunday = 7
func LocalWeekdayFromTimestamp(timestamp int64, region string) (int32, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0, err
	}
	dt := time.Unix(timestamp, 0)
	weekDay := dt.In(loc).Weekday()
	if weekDay == time.Sunday {
		return 7, nil
	}
	return int32(weekDay), nil
}

// LocalIsNeedRestDailyAtXHour : reset daily at xx hour
// Caution : return true if last update time is 0
func LocalIsNeedRestDailyAtXHour(resetHour int64, lastUpdateTimestamp int64, region string) (bool, error) {
	if lastUpdateTimestamp == 0 {
		return true, nil
	}
	nowTime := time.Now().Unix()
	offset := resetHour * 3600
	localDate, err := LocalDateFromTimestamp(nowTime-offset, region)
	if err != nil {
		return false, err
	}
	lastUpdateDate, err := LocalDateFromTimestamp(lastUpdateTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	if localDate != lastUpdateDate {
		return true, nil
	}
	return false, nil
}

// LocalIsNeedRestMonthlyAtXHour : reset monthly at xx hour
// Caution : return true if last update time is 0
func LocalIsNeedRestMonthlyAtXHour(resetHour int64, lastUpdateTimestamp int64, region string) (bool, error) {
	if lastUpdateTimestamp == 0 {
		return true, nil
	}
	nowTime := time.Now().Unix()
	offset := resetHour * 3600
	localDate, err := LocalMonthFromTimestamp(nowTime-offset, region)
	if err != nil {
		return false, err
	}
	lastUpdateDate, err := LocalMonthFromTimestamp(lastUpdateTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	if localDate != lastUpdateDate {
		return true, nil
	}
	return false, nil
}

// LocalIsNeedRestWeeklyAtXHour : reset monthly at xx hour
// Caution : return true if last update time is 0
func LocalIsNeedRestWeeklyAtXHour(resetHour int64, lastUpdateTimestamp int64, region string) (bool, error) {
	if lastUpdateTimestamp == 0 {
		return true, nil
	}
	nowTime := time.Now().Unix()
	offset := resetHour * 3600
	localDate, err := LocalMondayFromTimeStamp(nowTime-offset, region)
	if err != nil {
		return false, err
	}
	lastUpdateDate, err := LocalMondayFromTimeStamp(lastUpdateTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	if localDate != lastUpdateDate {
		return true, nil
	}
	return false, nil
}

// LocalIsNeedResetDailyAtXHourChkByNow : reset daily at xx hour
// Caution : return true if last update time is 0
func LocalIsNeedResetDailyAtXHourChkByNow(resetHour, nowTimestamp, lastUpdateTimestamp int64, region string) (bool, error) {
	if lastUpdateTimestamp == 0 {
		return true, nil
	}
	if nowTimestamp < lastUpdateTimestamp {
		return false, nil
	}
	offset := resetHour * 3600
	localDate, err := LocalDateFromTimestamp(nowTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	lastUpdateDate, err := LocalDateFromTimestamp(lastUpdateTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	if localDate == lastUpdateDate {
		return false, nil

	}
	return true, nil
}

// LocalIsNeedResetMonthlyAtXHourChkByNow : reset monthly at xx hour
// Caution : return true if last update time is 0
func LocalIsNeedResetMonthlyAtXHourChkByNow(resetHour, nowTimestamp, lastUpdateTimestamp int64, region string) (bool, error) {
	if lastUpdateTimestamp == 0 {
		return true, nil
	}
	if nowTimestamp < lastUpdateTimestamp {
		return false, nil
	}
	offset := resetHour * 3600
	localDate, err := LocalMonthFromTimestamp(nowTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	lastUpdateDate, err := LocalMonthFromTimestamp(lastUpdateTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	if localDate == lastUpdateDate {
		return false, nil
	}
	return true, nil
}

// LocalIsNeedResetWeeklyAtXHourChkByNow : reset monthly at xx hour
// Caution : return true if last update time is 0
func LocalIsNeedResetWeeklyAtXHourChkByNow(resetHour, nowTimestamp, lastUpdateTimestamp int64, region string) (bool, error) {
	if lastUpdateTimestamp == 0 {
		return true, nil
	}
	if nowTimestamp < lastUpdateTimestamp {
		return false, nil
	}
	offset := resetHour * 3600
	localDate, err := LocalMondayFromTimeStamp(nowTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	lastUpdateDate, err := LocalMondayFromTimeStamp(lastUpdateTimestamp-offset, region)
	if err != nil {
		return false, err
	}
	if localDate == lastUpdateDate {
		return false, nil
	}
	return true, nil
}

// LocalMondayFromTimeStamp return region Monday date, format: 2006-01-02
func LocalMondayFromTimeStamp(timestamp int64, region string) (string, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return "", err
	}
	dt := time.Unix(timestamp, 0).In(loc)
	weekDay := dt.Weekday()
	if weekDay == time.Sunday {
		dt = dt.AddDate(0, 0, -6)
	} else {
		dt = dt.AddDate(0, 0, -int(weekDay)+1)
	}
	return dt.Format("2006-01-02"), nil
}

// LocalYesterdayXHourTimeUnix using regional timezone
func LocalYesterdayXHourTimeUnix(addHour time.Duration, region string) int64 {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0
	}
	dateStr := time.Now().In(loc).Format("2006-01-02")
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return 0
	}
	t2 := t.Add(time.Hour * (-24 + addHour))
	return t2.Unix()
}

// LocalTodayXHourTimeUnix using regional timezone
func LocalTodayXHourTimeUnix(addHour time.Duration, region string) int64 {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0
	}
	dateStr := time.Now().In(loc).Format("2006-01-02")
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return 0
	}
	t2 := t.Add(time.Hour * addHour)
	return t2.Unix()
}

// LocalTomorrowXHourTimeUnix using regional timezone
func LocalTomorrowXHourTimeUnix(addHour time.Duration, region string) int64 {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0
	}
	dateStr := time.Now().In(loc).Format("2006-01-02")
	t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		return 0
	}
	t2 := t.Add(time.Hour * (24 + addHour))
	return t2.Unix()
}

func convLocalDateTime(datetime string, region string, format string) (string, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return "", err
	}
	dt, err := time.ParseInLocation("2006-01-02 15:04:05", datetime, time.Local)
	if err != nil {
		return "", err
	}
	return dt.In(loc).Format(format), nil
}

func convLocalNowDateTime(region string, format string) (string, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return "", err
	}
	return time.Now().In(loc).Format(format), nil
}

func convLocalDateTimeFromTimeStamp(timestamp int64, region string, format string) (string, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return "", err
	}
	dt := time.Unix(timestamp, 0)
	return dt.In(loc).Format(format), nil
}

// IsTheSameDay check two timestamp whether is the same day
func IsTheSameDay(newTimestamp int64, oldTimestamp int64, offsetSeconds int64, region string) (bool, error) {
	newDate, err := LocalDateFromTimestamp(newTimestamp-offsetSeconds, region)
	if err != nil {
		return false, err
	}
	oldDate, err := LocalDateFromTimestamp(oldTimestamp-offsetSeconds, region)
	if err != nil {
		return false, err
	}
	if newDate == oldDate {
		return true, nil
	}
	return false, nil
}

// IsTheSameDayByHour use offsetHour as the dividing time point, check two timestamp whether is the same day
func IsTheSameDayByHour(newTimestamp int64, oldTimestamp int64, offsetHour int64, region string) (bool, error) {
	offsetSeconds := offsetHour * 3600
	return IsTheSameDay(newTimestamp, oldTimestamp, offsetSeconds, region)
}

// IsTheSameWeek check two timestamp whether is the same week
func IsTheSameWeek(newTimestamp, oldTimestamp, offsetSeconds int64, region string) (bool, error) {
	newDate, err := LocalMondayFromTimeStamp(newTimestamp-offsetSeconds, region)
	if err != nil {
		return false, err
	}
	oldDate, err := LocalMondayFromTimeStamp(oldTimestamp-offsetSeconds, region)
	if err != nil {
		return false, err
	}

	return newDate == oldDate, nil
}

// IsTheSameWeekByHour use offsetHour as the dividing time point, check two timestamp whether is the same week
func IsTheSameWeekByHour(newTimestamp, oldTimestamp, offsetHour int64, region string) (bool, error) {
	return IsTheSameWeek(newTimestamp, oldTimestamp, offsetHour*3600, region)
}

// RegionMondayXHourTimestamp return region Monday 0x:00:00 timestamp
func RegionMondayXHourTimestamp(region string, addHour uint32) (int64, error) {
	nowTime := time.Now().Unix()
	localDate, err := LocalMondayFromTimeStamp(nowTime, region)
	if err != nil {
		return 0, err
	}
	dateStr := localDate + " 00:00:00"
	mondayTimeStamp, err := LocalDateTimeToTimeStamp(dateStr, region)
	if err != nil {
		return 0, err
	}
	return mondayTimeStamp + int64(addHour)*secsPerHour, nil
}

// RegionPassedDayFromDateTime return passed day from datetime to nowtime
func RegionPassedDayFromDateTime(region string, datetime string) (uint32, error) {
	loc, err := RegionTimeZoneLocation(region)
	if err != nil {
		return 0, err
	}
	nowTime := time.Now().In(loc)
	dt, err := time.ParseInLocation("2006-01-02 15:04:05", datetime, loc)
	if err != nil {
		return 0, err
	}
	duration := nowTime.Sub(dt)
	return uint32(duration.Seconds() / secsPerDay), nil
}

// RegionNextDailyResetHourTimeUnix return next reset time unix using regional timezone
func RegionNextDailyResetHourTimeUnix(resetHour int64, region string) (int64, error) {
	nowTime := time.Now().Unix()
	offset := resetHour * 3600
	localDate, err := LocalDateFromTimestamp(nowTime-offset, region)
	if err != nil {
		return 0, err
	}
	res, err := LocalDateToTimeStamp(localDate, region)
	if err != nil {
		return 0, err
	}
	return res + offset + secsPerDay, nil
}

// RegionNextDailyResetHourTimeUnixFrom : return next resetHour timestamp from given time
func RegionNextDailyResetHourTimeUnixFrom(timestamp int64, resetHour int64, region string) (int64, error) {
	offset := resetHour * 3600
	localDate, err := LocalDateFromTimestamp(timestamp-offset, region)
	if err != nil {
		return 0, err
	}
	res, err := LocalDateToTimeStamp(localDate, region)
	if err != nil {
		return 0, err
	}
	return res + offset + secsPerDay, nil
}

// The code below is obsolete, just for reference:

// var regionTimeZone = map[string]string{
// 	"VN":     "Asia/Ho_Chi_Minh",
// 	"TH":     "Asia/Bangkok",
// 	"ID":     "Asia/Jakarta",
// 	"TW":     "Asia/Taipei",
// 	"BR":     "America/Sao_Paulo",
// 	"SG":     "Asia/Singapore",
// 	"US":     "America/New_York",
// 	"MX":     "America/Mexico_City",
// 	"KR":     "Asia/Seoul",
// 	"RU":     "Europe/Moscow",
// 	"EUROPE": "Europe/Berlin",
// 	"CN":     "Asia/Shanghai",
// 	"ME":     "Asia/Dubai",
// 	"SAC":    "America/Sao_Paulo",
// 	"IND":    "Asia/Kolkata",
// }

// regionTimeZoneLocation : return time zone location
// func regionTimeZoneLocation(region string) (*time.Location, error) {
// 	loadCache := func() *time.Location {
// 		regionTimeZoneMux.RLock()
// 		defer regionTimeZoneMux.RUnlock()
// 		if loc, ok := regionTimeZoneLoc[region]; ok {
// 			return loc
// 		}
// 		return nil
// 	}

// 	loadFile := func() (*time.Location, error) {
// 		if _, ok := regionTimeZone[region]; !ok {
// 			region = defaultRegion
// 		}
// 		loc, err := time.LoadLocation(regionTimeZone[region])
// 		if err != nil {
// 			return nil, err
// 		}
// 		regionTimeZoneMux.Lock()
// 		defer regionTimeZoneMux.Unlock()
// 		regionTimeZoneLoc[region] = loc
// 		return loc, nil
// 	}

// 	if loc := loadCache(); loc != nil {
// 		return loc, nil
// 	}
// 	return loadFile()
// }
