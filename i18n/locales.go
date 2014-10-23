// 国际化操作.
package i18n

import (
	"encoding/json"
	"io/ioutil"
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
	// TODO: 自动加载所有语言配置

	load("zh_CN")
	load("en_US")
}

func load(localeStr string) {
	bytes, err := ioutil.ReadFile("i18n/" + localeStr + ".json")
	if nil != err {
		glog.Error(err)

		os.Exit(-1)
	}

	l := locale{Name: localeStr}

	err = json.Unmarshal(bytes, &l.Langs)
	if nil != err {
		glog.Error(err)

		os.Exit(-1)
	}

	Locales[localeStr] = l

	glog.V(5).Infof("Loaded [%s] locale configuration", localeStr)
}

// 获取语言配置项.
func Get(locale, key string) interface{} {
	return Locales[locale].Langs[key]
}

// 获取语言配置.
func GetAll(locale string) map[string]interface{} {
	return Locales[locale].Langs
}
