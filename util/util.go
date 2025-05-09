package util

import (
	"log"
	"os"
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

func SetupLogging() *os.File {
	val, ok := os.LookupEnv("TFINLINE_LOG_FILE")
	var logFile *os.File
	if !ok {
		val = os.DevNull // will this work on windows?
	}
	logFile, err := os.OpenFile(val, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	log.SetOutput(logFile)

	return logFile
}
