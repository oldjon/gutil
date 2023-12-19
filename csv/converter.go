package csv

import (
	"math"
	"strconv"
	"strings"
)

func StrToUint32(str string) uint32 {
	str = strings.ToLower(str)
	if str == "-inf" {
		return 0
	} else if str == "+inf" || str == "inf" {
		return math.MaxUint32
	}
	n, _ := strconv.ParseUint(str, 10, 32)
	return uint32(n)
}

func StrToUint64(str string) uint64 {
	str = strings.ToLower(str)
	if str == "-inf" {
		return 0
	} else if str == "+inf" || str == "inf" {
		return math.MaxUint64
	}
	n, _ := strconv.ParseUint(str, 10, 64)
	return n
}

func StrToInt32(str string) int32 {
	if str == "-inf" {
		return math.MinInt32
	} else if str == "+inf" || str == "inf" {
		return math.MaxInt32
	}
	n, _ := strconv.ParseInt(str, 10, 32)
	return int32(n)
}

func StrToInt64(str string) int64 {
	if str == "-inf" {
		return math.MinInt64
	} else if str == "+inf" || str == "inf" {
		return math.MaxInt64
	}
	n, _ := strconv.ParseInt(str, 10, 64)
	return n
}

func StrToFloat32(str string) float32 {
	n, _ := strconv.ParseFloat(str, 32)
	return float32(n)
}

func StrToFloat64(str string) float64 {
	n, _ := strconv.ParseFloat(str, 64)
	return n
}

func StrToBool(str string) bool {
	str = strings.ToLower(str)
	if str == "true" || str == "1" {
		return true
	}
	return false
}
