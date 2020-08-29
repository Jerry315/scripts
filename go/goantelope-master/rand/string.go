package rand

import (
	"math/rand"
	"time"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

// String 生成固定长度的随机字符串字符串
func String(n int) string {
	return StringWithLetters(letterBytes, n)
}

// StringWithLetters 生成固定长度的随机字符串字符串, 调用方提供字符列表
// ref: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func StringWithLetters(letters string, n int) string {
	rand.Seed(time.Now().UnixNano())
	buf := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letters) {
			buf[i] = letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(buf)
}
