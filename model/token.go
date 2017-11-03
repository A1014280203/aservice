package model

import (
	"time"
	"encoding/base64"
	"strings"
	"crypto/hmac"
	"crypto/sha256"
	"bytes"
)

type KeyCoup struct {
	// 手机号
	uid string
	secret []byte
	createdAt time.Time
}

// 这里实验一下，未绑定数组可以用吗？不可用代表猜测正确
func (kc *KeyCoup) checkData(data string) bool {
	parts := strings.Split(data, ".")
	var be base64.Encoding
	hb, err := be.DecodeString(parts[0])
	if err != nil {
		return false
	}
	pb, err := be.DecodeString(parts[0])
	if err != nil {
		return false
	}
	sb, err := be.DecodeString(parts[0])
	if err != nil {
		return false
	}
	
	h := hmac.New(sha256.New, kc.secret)
	_, err = h.Write(hb)
	if err != nil {
		return false
	}
	_, err = h.Write(pb)
	if err != nil {
		return false
	}
	
	if bytes.Compare(h.Sum(nil), sb) != 0 {
		return false
	}
	return true
}

type KeyCoupManager struct {
	storage map[string]*KeyCoup
	// nano
	ttl int64
}

func genSecret() []byte {
	// buf := make([]byte, 16)
	// if _, err := io.ReadFull(rand.Reader, buf); err!=nil {
	// 	return ""
	// }
	// return base64.URLEncoding.EncodeToString(buf)
	return genRandomByte(16)
}

// 复用会话需要重新创建KeyCoup
func (km* KeyCoupManager) CreateKeyCoup(uid string) *KeyCoup {
	var kc KeyCoup
	kc.uid = uid
	kc.secret = genSecret()
	kc.createdAt = time.Now()
	km.storage[uid] = &kc
	return &kc
}

func (km *KeyCoupManager)isExpired(kc *KeyCoup) bool {
	if km.ttl < time.Now().Unix() - kc.createdAt.Unix() {
		return true
	}
	return false
}

func (km *KeyCoupManager) Valid(uid string, data string) bool {
	scrt := km.storage[uid]
	if scrt == nil {
		return false
	}
	if km.isExpired(scrt) {
		return false
	}
	return scrt.checkData(data)
}

func InitialKeyCoupManager(ttlSec int64) KeyCoupManager {
	var km KeyCoupManager
	km.storage = make(map[string]*KeyCoup)
	km.ttl = ttlSec * 1000 * 1000 * 1000
	return km
}

