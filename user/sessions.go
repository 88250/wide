package user

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
)

const (
	SessionStateActive = iota // 会话状态：活的
)

// 用户 HTTP 会话，用于验证登录.
var Session = sessions.NewCookieStore([]byte("BEYOND"))

// Wide 会话，对应一个浏览器 tab.
type WideSession struct {
	Id      string    // 唯一标识
	State   int       // 状态
	Created time.Time // 创建时间
	Updated time.Time // 最近一次使用时间
}

// 创建一个 Wide 会话.
func NewSession() *WideSession {
	rand.Seed(time.Now().UnixNano())

	id := strconv.Itoa(rand.Int())
	now := time.Now()

	return &WideSession{
		Id:      id,
		State:   SessionStateActive,
		Created: now,
		Updated: now,
	}
}
