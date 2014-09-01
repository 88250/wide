package util

import (
	"encoding/json"
	"github.com/golang/glog"
	"net/http"
)

func RetJSON(w http.ResponseWriter, r *http.Request, res map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(res)
	if err != nil {
		glog.Error(err)
		return
	}

	w.Write(data)
}
