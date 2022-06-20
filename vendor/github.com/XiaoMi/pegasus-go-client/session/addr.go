package session

import (
	"fmt"
	"net"
)

// ResolveMetaAddr into a list of TCP4 addresses. Error is returned if the given `addrs` are not either
// a list of valid TCP4 addresses, or a resolvable hostname.
func ResolveMetaAddr(addrs []string) ([]string, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("meta server list should not be empty")
	}

	// case#1: all addresses are in TCP4 already
	allTCPAddr := true
	for _, addr := range addrs {
		_, err := net.ResolveTCPAddr("tcp4", addr)
		if err != nil {
			allTCPAddr = false
			break
		}
	}
	if allTCPAddr {
		return addrs, nil
	}

	// case#2: address is a hostname
	if len(addrs) == 1 {
		actualAddrs, err := net.LookupHost(addrs[0])
		if err == nil {
			return actualAddrs, nil
		}
	}

	return nil, fmt.Errorf("illegal meta addresses: %s", addrs)
}
