package model

import (
	"encoding/json"
	"gos/dbc"
	"time"
	"io"
	"crypto/rand"
	"encoding/base64"
)

func init() {
	Sessions = createSessionManager()
	go Sessions.sessionHunter()
}

var Sessions SessionManager

type Session struct {
	sid        string
	buf        map[string]string
	createdAt  time.Time
	lastAccess time.Time
	state      int8 // -1->不可用/请求被销毁
}

func (s *Session) updateAccessRecord() {
	s.lastAccess = time.Now()
}

func genSessionID() string {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err!=nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf)
}

// GetSessionID 返回会话ID
// 目前设计的sid对使用者是“只读”
func (s *Session) GetSessionID() string {
	return s.sid
}

// ToJSON return Encoded buf filed
func (s *Session) ToJSON() []byte {
	b, _ := json.Marshal(s.buf)
	s.updateAccessRecord()
	return b
}

// func (s *Session) SetSessionID(sid string) {
// 	s.sid = sid
// 	s.updateAccessRecord()
// }

func (s *Session) SetKV(k, v string) {
	s.buf[k] = v
	s.updateAccessRecord()
}

func (s *Session) GetValue(k string) string {
	s.updateAccessRecord()
	return s.buf[k]
}

func (s *Session) CleanBuf() {
	s.buf = make(map[string]string)
	s.updateAccessRecord()
}

func (s *Session) DestroyMe() {
	s.state = -1
}

// SessionManager 会话管理，绝大多数时候应该通过这个操作会话
type SessionManager struct {
	// sessionStorage []Session
	storage  map[string]*Session
	// size     int
	// rend     int
	deltaSec int64
	ttl      int
}

// SetDeltaSec 设置无访问会话在内存中存在的最长时间
func (sm *SessionManager) SetDeltaSec(sec int64) {
	sm.deltaSec = sec
}

// SetTTL 设置无访问会话在redis中的存续时间
func (sm *SessionManager) SetTTL(sec int) {
	sm.ttl = sec
}
// 检查是否过长时间未使用
func (sm *SessionManager) isDisrepaired(s *Session) bool {
	if time.Now().Unix()-s.lastAccess.Unix() > sm.deltaSec {
		return true
	}
	return false
}

func (sm *SessionManager) isDamaged(s *Session) bool {
	if s.state == -1 {
		return true
	}
	return false
}

// 删去内存中sid为所给值的会话
func (sm *SessionManager) subStorage(sid string) {
	// 下面是会话存储列表实现时的删减操作
	// 这里申请一个和添加会话互斥的锁
	// var i int
	// for i = 0; i < len(sm.sessionStorage); i++ {
	// 	if sm.sessionStorage[i].sid == sid {
	// 		break
	// 	}
	// }
	// if i >= len(sm.sessionStorage) {
	// 	return
	// } else if i == 0 {
	// 	sm.sessionStorage = sm.sessionStorage[1:]
	// } else if i == len(sm.sessionStorage)-1 {
	// 	sm.sessionStorage = sm.sessionStorage[:len(sm.sessionStorage)-1]
	// } else {
	// 	sm.sessionStorage = append(sm.sessionStorage[:i], sm.sessionStorage[i+1:]...)
	// }
	// sm.rend--
	// 这里释放锁
	delete(sm.storage, sid)
}

// 序列化并传到redis s.sid=>s.buf
func (sm *SessionManager) submit(s *Session) {
	b, _ := json.Marshal(s.ToJSON())
	// 这里应该有过期时间
	dbc.SetKeyByteValue(s.sid, b, sm.ttl)
}

// 将会话放到内存
func (sm *SessionManager) push(sesp *Session) *Session {
	// 下面是会话存储列表实现时的存储操作
	// 这里要申请一个和删减函数互斥的锁
	// sm.sessionStorage[sm.rend] = ses
	// _sp := &sm.sessionStorage[sm.rend]
	// sm.rend++
	// 这里释放锁
	sm.storage[sesp.sid] = sesp
	return sesp
}

// SessionHunter 监视内存中存储的会话情
// 将不常用的会话放在redis并从内存中删去
// 将请求销毁的会话从内存中删去
// 这应该是个独立函数，注意互斥
// 会出现的问题，无非就是正在被使用的被删除，所以对会话的取用都要通过函数操作；包括对会话的取用
func (sm *SessionManager) sessionHunter() {
	for _, sesp := range sm.storage {
		if sm.isDisrepaired(sesp) || sm.isDamaged(sesp) {
			defer sm.subStorage(sesp.sid)
			sm.submit(sesp)
		}
	}
}

// GetSessionBySid 内存->redis->nil
func (sm *SessionManager) GetSessionBySid(sid string) *Session {
	// 从内存中寻找
	sesp := sm.getFromMemo(sid)
	if sesp != nil {
		return sesp
	}
	// 从redis寻找
	return sm.getFromRemote(sid)
}

// 从内存中寻找会话
func (sm *SessionManager) getFromMemo(sid string) *Session {
	sesp := sm.storage[sid]
	if sesp == nil || sesp.state == -1 {
		return nil
	}
	return sesp
}

// 从redis取回并实例化
func (sm *SessionManager) getFromRemote(sid string) *Session {
	b, err := dbc.GetKeyByteValue(sid)
	if err == nil {
		if string(b) != "" {
			var _buf map[string]string
			json.Unmarshal(b, &_buf)
			var ses Session
			ses.buf = _buf
			ses.sid = sid
			ses.createdAt = time.Now()
			return sm.push(&ses)
		}
	}
	return nil
}

// Save 本地应该是没有大小限制的
// 这样，将长时间不用的，传到redis
// 在redis上设置过期时间
// 当新的数据推送到redis的时候，过期时间被重设
// 存储会话
func (sm *SessionManager) Save(s *Session) {
	sm.push(s)
}

// createSessionManager 启动会话管理程序
// 初始化结构体中的各个参数
func createSessionManager() SessionManager {
	var sm SessionManager
	sm.storage = make(map[string]*Session)
	sm.deltaSec = 3600 * 24
	sm.ttl = 3600 * 24 * 2
	return sm
}

// CreateSession 创建会话，并返回会话指针
func (sm *SessionManager) CreateSession() *Session {
	var s Session
	s.sid = genSessionID()
	s.createdAt = time.Now()
	// 为什么有的时候需要初始化？
	s.buf = make(map[string]string)
	sm.push(&s)
	return &s
}
