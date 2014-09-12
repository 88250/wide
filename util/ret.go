package util

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// HTTP 返回 JSON 统一处理.
func RetJSON(w http.ResponseWriter, r *http.Request, res map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(res)
	if err != nil {
		glog.Error(err)
		return
	}

	w.Write(data)
}
