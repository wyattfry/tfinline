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

type TerraformObject struct {
	Address, Id string
}

func ExtractResourceAddressAndId(line string) *TerraformObject {
	parts := strings.Split(line, "\"")
	if len(parts) > 2 {
		return &TerraformObject{
			Address: strings.TrimSpace(parts[1]),
			Id:      strings.TrimSpace(parts[3]),
		}
	}
	return nil
}
