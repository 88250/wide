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

// Package i18n includes internationalization related manipulations.
package i18n

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/b3log/wide/log"
)

// Logger.
var logger = log.NewLogger(os.Stdout)

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
		logger.Error(err)

		os.Exit(-1)
	}

	l := locale{Name: localeStr}

	err = json.Unmarshal(bytes, &l.Langs)
	if nil != err {
		logger.Error(err)

		os.Exit(-1)
	}

	Locales[localeStr] = l
}

// Get gets a message with the specified locale and key.
func Get(locale, key string) interface{} {
	return Locales[locale].Langs[key]
}

// GetAll gets all messages with the specified locale.
func GetAll(locale string) map[string]interface{} {
	return Locales[locale].Langs
}

// GetLocalesNames gets names of all locales. Returns ["zh_CN", "en_US"] for example.
func GetLocalesNames() []string {
	ret := []string{}

	for name := range Locales {
		ret = append(ret, name)
	}

	sort.Strings(ret)

	return ret
}
