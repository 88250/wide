package util

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket channel.
type WSChannel struct {
	Sid     string          // wide session id
	Conn    *websocket.Conn // websocket connection
	Request *http.Request   // HTTP request related
	Time    time.Time       // the latest use time
}

// Close closed the channel.
func (c *WSChannel) Close() {
	c.Conn.Close()
}

// Refresh refreshes the channel by updating its use time.
func (c *WSChannel) Refresh() {
	c.Time = time.Now()
}
