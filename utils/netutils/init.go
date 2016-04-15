package netutils

import (
	"errors"
	"fmt"
	"strings"
)

const (
	networkDelimiter = "@"
)

// NetAddr combine network and laddr.
type NetAddr struct {
	Network string
	LAddr   string
}

// ParseNetAddr parse network@laddr from netAddr str.
func ParseNetAddr(str string) (netAddr *NetAddr, err error) {
	idx := strings.Index(str, networkDelimiter)
	if idx == -1 {
		err = errors.New("parse error")
		return
	}

	netAddr = &NetAddr{
		Network: str[:idx],
		LAddr:   str[idx+1:],
	}
	return
}

func (netAddr *NetAddr) String() string {
	return fmt.Sprintf("%s@%s", netAddr.Network, netAddr.LAddr)
}
