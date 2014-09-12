package user

import (
	"github.com/gorilla/sessions"
)

// 用户会话.
var Session = sessions.NewCookieStore([]byte("BEYOND"))
