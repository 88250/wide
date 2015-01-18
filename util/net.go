// Copyright (c) 2014-2015, b3log.org
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

// Package util includes common utilities.
package util

import (
	"errors"
	"net"
)

type mynet struct{}

// Network utilities.
var Net = mynet{}

// LocalIP gets the first NIC's IP address.
func (*mynet) LocalIP() (string, error) {
	tt, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	for _, t := range tt {
		aa, err := t.Addrs()

		if err != nil {
			return "", err
		}

		for _, a := range aa {
			switch ip := a.(type) {
			case *net.IPAddr:
				v4 := ip.IP.To4()

				if v4 == nil || v4.IsLoopback() || v4.IsUnspecified() {
					continue
				}

				return v4.String(), nil
			case *net.IPNet:
				v4 := ip.IP.To4()

				if v4 == nil || v4.IsLoopback() || v4.IsUnspecified() {
					continue
				}

				return v4.String(), nil
			default:
				return "", errors.New("cannot find local IP address")
			}
		}
	}

	return "", errors.New("cannot find local IP address")
}
