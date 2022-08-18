//Package utils provides helper functions for all app.
package utils

import (
	"crypto/md5"
	"fmt"
)

//MD5 generates a fixed length hash from a string
func MD5(data string) string {
	h := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", h)
}
