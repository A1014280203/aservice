package model

import (
	"encoding/base64"
	"io"
	"crypto/rand"
)

// 没有放置错误处理，需要判断可用性有多少
func genRandomByte(n int) ([]byte) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, buf); err!=nil {
		return []byte{}
	}
	return buf
}

func genRandomBase64Str(n int) string {
	b := genRandomByte(n)
	return base64.URLEncoding.EncodeToString(b)
}


func base64Decode(str string) []byte {
	b, _ := base64.URLEncoding.DecodeString(str)
	return b
}