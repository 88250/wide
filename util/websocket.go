// Copyright (c) 2014-2016, b3log.org & hacpai.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WSChannel represents a WebSocket channel.
type WSChannel struct {
	Sid     string          // wide session id
	Conn    *websocket.Conn // websocket connection
	Request *http.Request   // HTTP request related
	Time    time.Time       // the latest use time
}

// WriteJSON writes the JSON encoding of v to the channel.
func (c *WSChannel) WriteJSON(v interface{}) (ret error) {
	if nil == c.Conn {
		return errors.New("connection is nil, channel has been closed")
	}

	defer func() {
		if r := recover(); nil != r {
			ret = errors.New("channel has been closed")
		}
	}()

	return c.Conn.WriteJSON(v)
}

// ReadJSON reads the next JSON-encoded message from the channel and stores it in the value pointed to by v.
func (c *WSChannel) ReadJSON(v interface{}) (ret error) {
	if nil == c.Conn {
		return errors.New("connection is nil, channel has been closed")
	}

	defer func() {
		if r := recover(); nil != r {
			ret = errors.New("channel has been closed")
		}
	}()

	return c.Conn.ReadJSON(v)
}

// Close closed the channel.
func (c *WSChannel) Close() {
	if nil != c.Conn {
		c.Conn.Close()
	}
}

// Refresh refreshes the channel by updating its use time.
func (c *WSChannel) Refresh() {
	c.Time = time.Now()
}
