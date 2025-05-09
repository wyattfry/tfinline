package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wyattfry/tfinline/event"
	"github.com/wyattfry/tfinline/inline"
	"github.com/wyattfry/tfinline/util"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lines := make(chan string, 128)
	processedLines := make(chan string, 128) // Intermediate channel
	// Forward processed lines to the View
	go func() {
		for line := range processedLines {
			lines <- line // Forward processed lines to View
		}
	}()
	done := make(chan struct{}) // signals view is finished
	go inline.View(lines, done)

	cmdQueue := make([]exec.Cmd, 1)

	args := os.Args[1:]
	if args[0] != "import" {
		args = append(args, "-json")
	}
	firstCmd := exec.CommandContext(ctx, "terraform", args...)

	var mu sync.Mutex // Protects the queue
	mu.Lock()
	cmdQueue[0] = *firstCmd
	mu.Unlock()

	for len(cmdQueue) > 0 {
		mu.Lock()
		cmd := cmdQueue[0]
		cmdQueue = cmdQueue[1:]
		mu.Unlock()

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)
		go processOutput(&wg, stdout, &cmdQueue, &mu, processedLines, ctx, cmd)
		go processOutput(&wg, stderr, &cmdQueue, &mu, processedLines, ctx, cmd)
		wg.Wait()
		if err := cmd.Wait(); err != nil {
			log.Println("the command failed to run or didn't complete successfully:", cmd.String(), err)
		}
	}

	close(processedLines)
	close(lines) // Close the channel after all commands are processed
	<-done       // Wait for inline.View() to finish
}

func processOutput(wg *sync.WaitGroup, r io.Reader, queue *[]exec.Cmd, mu *sync.Mutex, processedLines chan<- string, ctx context.Context, command exec.Cmd) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "error: A resource with the ID") {
			tfobj := util.ExtractResourceAddressAndId(line)
			importCmd := exec.CommandContext(ctx, "terraform", "import", "-no-color", tfobj.Address, tfobj.Id)
			mu.Lock()
			*queue = append(*queue, *importCmd)
			mu.Unlock()
		}

		if command.Args[1] == "import" {
			// Import doesn't support JSON output, so we fake it
			e := event.Event{
				Message: line,
				Diagnostic: &event.Diagnostic{
					Address: command.Args[3],
				},
				Level: "info",
				Type:  event.ImportSomething,
			}
			byteArrayLine, _ := json.Marshal(e)
			line = string(byteArrayLine)
		}

		processedLines <- line // Send processed line to intermediate channel
	}
}
