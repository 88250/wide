// Internationalization manipulations.
package i18n

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
)

// Locale.
type locale struct {
	Name     string
	Langs    map[string]interface{}
	TimeZone string
}

// All locales.
var Locales = map[string]locale{}

// Load loads i18n message configurations.
func Load() {
	f, _ := os.Open("i18n")
	names, _ := f.Readdirnames(-1)
	f.Close()

	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		loc := name[:strings.LastIndex(name, ".")]
		load(loc)
	}
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

// Get gets message with the specified locale and key.
func Get(locale, key string) interface{} {
	return Locales[locale].Langs[key]
}

// GetAll gets all messages with the specified locale.
func GetAll(locale string) map[string]interface{} {
	return Locales[locale].Langs
}
