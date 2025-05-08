package util

import (
	"regexp"
	"strings"
)

func TrimAddrPrefix(msg, addr string) string {
	msg = regexp.MustCompile(` \[id=.*\]`).ReplaceAllString(msg, "")
	if strings.HasPrefix(msg, addr+": ") {
		return msg[len(addr)+2:]
	}
	return msg
}
