package util

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// 一个用户会话的 WebSocket 通道结构.
type WSChannel struct {
	Sid     string          // 用户会话 id
	Conn    *websocket.Conn // WebSocket 连接
	Request *http.Request   // 关联的 HTTP 请求
	Time    time.Time       // 该通道最近一次使用时间
}
