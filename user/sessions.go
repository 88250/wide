package user

import (
	"github.com/gorilla/sessions"
)

var Session = sessions.NewCookieStore([]byte("BEYOND"))
