package util

import (
	"regexp"
	"strings"
)

// Remove resource name prefix from message and remove resource ID for brevity
// e.g. "azurerm_resource_group.example: Still updating [id="a-very-long-resource-id"]"
// becomes: "Still updating"
func TrimAddrPrefix(msg, addr string) string {
	msg = regexp.MustCompile(` \[id=.*\]`).ReplaceAllString(msg, "")
	if strings.HasPrefix(msg, addr+": ") {
		return msg[len(addr)+2:]
	}
	return msg
}
