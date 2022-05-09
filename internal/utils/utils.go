package utils

import (
	"crypto/md5"
	"fmt"
)

func MD5(data string) string {
	h := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", h)
}
