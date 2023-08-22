package crypto

import (
	"crypto/md5"
	"fmt"
)

func MD5Sum(in []byte) []byte {
	out := md5.Sum(in)
	return out[:]
}

func MD5SumStr(in string) string {
	out := md5.Sum([]byte(in))
	return fmt.Sprintf("%x", out)
}
