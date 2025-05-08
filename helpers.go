package main

import (
	"log"
	"os"
)

func setupLogging() *os.File {
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
