// Copyright (c) 2014-present, b3log.org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/88250/gulu"
	"github.com/parnurzeal/gorequest"
)

// Logger
var logger = gulu.Log.NewLogger(os.Stdout)

// HacPaiURL is the URL of HacPai community.
const HacPaiURL = "https://ld246.com"

// HacPaiUserInfo returns HacPai community user info specified by the given access token.
func HacPaiUserInfo(accessToken string) (ret map[string]interface{}) {
	result := map[string]interface{}{}
	response, data, errors := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Post(HacPaiURL+"/user/ak").SendString("access_token="+accessToken).Timeout(7*time.Second).
		Set("User-Agent", "Pipe; +https://github.com/88250/wide").EndStruct(&result)
	if nil != errors || http.StatusOK != response.StatusCode {
		logger.Errorf("get community user info failed: %+v, %s", errors, data)
		return nil
	}
	if 0 != result["code"].(float64) {
		return nil
	}
	return result["data"].(map[string]interface{})
}
