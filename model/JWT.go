package model

import (
	"encoding/json"
	"encoding/base64"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"
)

func init() {
	
	JWTManager.secret = genRandomByte(16)
	JWTManager.alg = "sha256"
	JWTManager.typ = "JWT"
}

var JWTManager JWT

// 暂时都使用sha256
type JWT struct {
	// header
	alg string
	typ string
	// 密钥
	secret []byte

}

// SetAlg 暂时并不使用
func (jwt *JWT)SetAlg(alg string) {
	jwt.alg = alg
}

func (jwt *JWT) hmac_SHA256(args... []byte) []byte {
	h := hmac.New(sha256.New, jwt.secret)
	for _, b := range args {
		h.Write(b)
	}
	return h.Sum(nil)
}

func (jwt *JWT) ValidJWT(token string) bool {
	t := strings.Split(token, ".")
	if len(t) != 3 {
		return false
	}
	// header, payload, signature in string
	hstr, pstr, sstr := t[0], t[1], t[2]
	// header, payload, signature in bytes
	hb := base64Decode(hstr)
	pb := base64Decode(pstr)
	sb := base64Decode(sstr)
	if bytes.Compare(sb, jwt.hmac_SHA256(hb, pb)) == 0 {
		return true
	}
	return false
}

// MakeJWT args must can be JSON Encode
func (jwt *JWT) MakeJWT(hm map[string]string, pm interface{}) string {
	// header bytes
	hm["alg"] = jwt.alg
	hm["typ"] = jwt.typ
	hb, _ := json.Marshal(hm)
	// for test
	// hb1, err := json.Marshal(hm)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// payload bytes
	pb, _ := json.Marshal(pm)
	// signature bytes
	sb := jwt.hmac_SHA256(hb, pb)
	// base64 string
	hstr := base64.URLEncoding.EncodeToString(hb)
	pstr := base64.URLEncoding.EncodeToString(pb)
	sstr := base64.URLEncoding.EncodeToString(sb)
	return fmt.Sprintf("%s.%s.%s", hstr, pstr, sstr)
}

