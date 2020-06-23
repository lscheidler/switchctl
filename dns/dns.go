package dns

import (
	"net"
)

func Check(hostname string) bool {
	if _, err := net.LookupHost(hostname); err != nil {
		return false
	} else {
		return true
	}
}
