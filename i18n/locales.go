// 国际化操作.
package i18n

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/glog"
)

type locale struct {
	Name     string
	Langs    map[string]interface{}
	TimeZone string
}

// 所有的 locales.
var Locales = map[string]locale{}

// 加载国际化配置.
func Load() {
	// TODO: 加载所有语言配置
	bytes, _ := ioutil.ReadFile("i18n/zh_CN.json")

	zhCN := locale{Name: "zh_CN"}

	// TODO: 时区

	err := json.Unmarshal(bytes, &zhCN.Langs)
	if err != nil {
		glog.Error(err)

		os.Exit(-1)
	}

	Locales["zh_CN"] = zhCN
	glog.V(5).Info("Loaded [zh_CN] locale configuration")
}

// 获取请求对应的本地语言配置.
func GetLangs(r *http.Request) map[string]interface{} {
	locale := GetLocale(r)

	return Locales[locale].Langs
}

// 获取请求对应的 locale.
func GetLocale(r *http.Request) string {
	// TODO: 从请求中获取 locale
	return "zh_CN"
}
