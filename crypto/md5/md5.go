package md5

import (
	"crypto/md5"
	"fmt"
)

func Sum(in []byte) []byte {
	out := md5.Sum(in)
	return out[:]
}

func SumStr(in string) string {
	out := md5.Sum([]byte(in))
	return fmt.Sprintf("%x", out)
}
