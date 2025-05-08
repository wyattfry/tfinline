package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/wyattfry/tfinline/inline"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Diagnostic struct {
	Severity string `json:"severity"`
	Address  string `json:"address"`
	Detail   string `json:"detail"`
}

type ChangeSummary struct {
	Add       int    `json:"add"`
	Change    int    `json:"change"`
	Import    int    `json:"import"`
	Remove    int    `json:"remove"`
	Operation string `json:"operation"`
}

type Event struct {
	Level      string                 `json:"@level"`
	Type       string                 `json:"type"`
	Hook       map[string]interface{} `json:"hook"`
	Message    string                 `json:"@message"`
	Diagnostic *Diagnostic            `json:"diagnostic,omitempty"`
	Changes    *ChangeSummary         `json:"changes,omitempty"`
}

func main() {
	logFile := setupLogging()
	defer logFile.Close()

	log.Println("tfinline started")

	var tfSubCmd string
	var importArgs string

	flag.StringVar(&tfSubCmd, "cmd", "apply", "terraform command (apply|plan|destroy|init)")
	flag.StringVar(&importArgs, "import", "", "run terraform import (pass full import arguments)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nEnvironment Variables:")
		fmt.Fprintln(os.Stderr, "  TFINLINE_LOG_FILE  Optional Path to the log file for tfinline")
	}
	flag.Parse()

	args := os.Args[1:]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	args = append(args, "-json")

	cmd := exec.CommandContext(ctx, "terraform", args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return
	} else {
		fmt.Println("Refreshing State...")
	}

	lines := make(chan string, 128)
	done := make(chan struct{}) // signals view is finished

	var wg sync.WaitGroup
	wg.Add(2)
	go pipeToChan(&wg, stdout, lines)
	go pipeToChan(&wg, stderr, lines)

	go inline.View(lines, done)

	// wait for pipes, then for terraform itself
	wg.Wait()
	close(lines) // -> tells UI no more lines
	cmd.Wait()
	<-done // UI drained & bars closed
}

func pipeToChan(wg *sync.WaitGroup, r io.Reader, out chan<- string) {
	defer wg.Done()
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		out <- sc.Text()
	}
}

type resourceEvent struct {
	addr, action string
}
