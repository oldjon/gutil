package gutil

import (
	"crypto/md5"
	"encoding/hex"
	"golang.org/x/exp/constraints"
)

func IF[T interface{}](cdt bool, a, b T) T {
	if cdt {
		return a
	}
	return b
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
