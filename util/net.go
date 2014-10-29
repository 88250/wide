// Utilities.
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
