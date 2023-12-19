package gutil

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

func If[T any](ok bool, trueValue, falseValue T) T {
	if ok {
		return trueValue
	}
	return falseValue
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Bound[T constraints.Ordered](a, min, max T) T {
	if min > max {
		min, max = max, min
	}
	if a < min {
		return min
	} else if a > max {
		return max
	}
	return a
}

func StrToUint32(str string) uint32 {
	n, _ := strconv.ParseUint(str, 10, 32)
	return uint32(n)
}

func StrToUint64(str string) uint64 {
	n, _ := strconv.ParseUint(str, 10, 64)
	return n
}

func StrToInt32(str string) int32 {
	n, _ := strconv.ParseInt(str, 10, 32)
	return int32(n)
}

func StrToInt64(str string) int64 {
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

func Uint32ToString(i uint32) string {
	return strconv.FormatUint(uint64(i), 10)
}

func Uint64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func Int32ToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}
