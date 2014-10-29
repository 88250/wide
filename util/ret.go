package util

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// RetJSON writes HTTP response with "Content-Type, application/json".
func RetJSON(w http.ResponseWriter, r *http.Request, res map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(res)
	if err != nil {
		glog.Error(err)
		return
	}

	w.Write(data)
}
