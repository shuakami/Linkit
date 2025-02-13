package utils

import (
	"crypto/rand"
	"math/big"
	"regexp"
)

const (
	// 默认字符集
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// 自定义短码的正则表达式
	customCodePattern = "^[a-zA-Z0-9-_]{4,16}$"
)

// GenerateShortCode 生成随机短码
func GenerateShortCode(length int) (string, error) {
	b := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}

	return string(b), nil
}

// ValidateCustomCode 验证自定义短码
func ValidateCustomCode(code string) bool {
	match, _ := regexp.MatchString(customCodePattern, code)
	return match
}
