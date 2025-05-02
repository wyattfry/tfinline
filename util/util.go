package util

import (
	"regexp"
	"strings"
)

func TrimAddrPrefix(msg, addr string) string {

	// remove the resource ID from the message, they're long.
	msg = regexp.MustCompile(` \[id=.*\]`).ReplaceAllString(msg, "")

	if strings.HasPrefix(msg, addr+": ") {
		return msg[len(addr)+2:]
	}
	return msg
}
