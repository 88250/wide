package user

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/sessions"
)

const (
	SessionStateActive = iota // 会话状态：活的
)

// 用户 HTTP 会话，用于验证登录.
var HTTPSession = sessions.NewCookieStore([]byte("BEYOND"))

// Wide 会话，对应一个浏览器 tab.
type WideSession struct {
	Id            string    // 唯一标识
	HTTPSessionId string    // HTTP 会话 id
	State         int       // 状态
	Created       time.Time // 创建时间
	Updated       time.Time // 最近一次使用时间
}

type Sessions []*WideSession

// 所有 Wide 会话集.
var WideSessions Sessions

// 创建一个 Wide 会话.
func (sessions *Sessions) New() *WideSession {
	rand.Seed(time.Now().UnixNano())

	id := strconv.Itoa(rand.Int())
	now := time.Now()

	ret := &WideSession{
		Id:      id,
		State:   SessionStateActive,
		Created: now,
		Updated: now,
	}

	*sessions = append(*sessions, ret)

	return ret
}

// 移除 Wide 会话.
func (sessions *Sessions) Remove(sid string) {
	for i, s := range *sessions {
		if s.Id == sid {
			*sessions = append((*sessions)[:i], (*sessions)[i+1:]...)

			glog.V(3).Infof("Has [%d] wide sessions currently", len(*sessions))
		}
	}
}

// 获取 HTTP 会话关联的所有 Wide 会话.
func (sessions *Sessions) GetByHTTPSid(httpSessionId string) []*WideSession {
	ret := []*WideSession{}

	for _, s := range *sessions {
		if s.HTTPSessionId == httpSessionId {
			ret = append(ret, s)
		}
	}

	return ret
}

// 移除 HTTP 会话关联的所有 Wide 会话.
func (sessions *Sessions) RemoveByHTTPSid(httpSessionId string) {
	for i, s := range *sessions {
		if s.HTTPSessionId == httpSessionId {
			*sessions = append((*sessions)[:i], (*sessions)[i+1:]...)

			glog.V(3).Infof("Has [%d] wide sessions currently", len(*sessions))
		}
	}
}
