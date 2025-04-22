package main

import (
	"log"
	"os"
)

type Diagnostic struct {
	Severity string `json:"severity"`
}

// "changes":{"add":0,"change":0,"import":0,"remove":3,"operation":"plan"}
type ChangeSummary struct {
	Add       int    `json:"add"`
	Change    int    `json:"change"`
	Import    int    `json:"import"`
	Remove    int    `json:"remove"`
	Operation string `json:"operation"`
}

type Event struct {
	Type       string                 `json:"type"`
	Hook       map[string]interface{} `json:"hook"`
	Message    string                 `json:"@message"`
	Diagnostic *Diagnostic            `json:"diagnostic,omitempty"`
	Changes    *ChangeSummary         `json:"changes,omitempty"`
}

func main() {
	logFile, err := os.OpenFile("tfinline.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("tfinline started")

	args := os.Args[1:]
	pretty := isApplyOrDestroy(args)

	cmd := buildCmd(args, pretty)
	stdout, err := cmd.StdoutPipe()
	must(err)
	// cmd.Stderr = os.Stderr
	cmd.Stderr = log.Writer()
	must(cmd.Start())

	if pretty {
		runPretty(stdout)
	} else {
		passThrough(stdout)
	}

	must(cmd.Wait())
}
