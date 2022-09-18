//Package utils provides helper functions for all app.
package utils

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/zhel1/yandex-practicum-go/internal/dto"
)

//MD5 generates a fixed length hash from a string
func MD5(data string) string {
	h := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", h)
}

func ExtractValueFromContext(context context.Context, name interface{}) (string, error) {
	value := ""
	if id := context.Value(name); id != nil {
		value = id.(string)
	}

	if value == "" {
		return "", errors.New("empty " + name.(dto.UserConst).String()) //TODO
	}
	return value, nil
}
