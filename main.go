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

func main() {
	logFile := util.SetupLogging()
	defer logFile.Close()

	log.Println("tfinline started")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s [init|plan|apply|destroy] ...\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\nEnvironment Variables:")
		fmt.Fprintln(os.Stderr, "  TFINLINE_LOG_FILE  Optional Path to the log file for tfinline")
	}
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lines := make(chan string, 128)
	processedLines := make(chan string, 128) // Intermediate channel
	go func() {
		for line := range processedLines {
			lines <- line // Send processed lines to View
		}
	}()
	done := make(chan struct{}) // signals view is finished
	go inline.View(lines, done)

	// If an 'apply' command produces 'resource already exists' errors,
	// we queue up import commands to run after the apply command,
	// and run until the command queue is empty.
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
			tfObject := util.ExtractResourceAddressAndId(line)
			importCmd := exec.CommandContext(ctx, "terraform", "import", "-no-color", tfObject.Address, tfObject.Id)
			mu.Lock()
			*queue = append(*queue, *importCmd)
			mu.Unlock()
		}

		if command.Args[1] == "import" {
			// Import doesn't support JSON output, so we fake it
			byteArrayLine, _ := json.Marshal(event.Event{
				Message: line,
				Diagnostic: &event.Diagnostic{
					Address: command.Args[3],
				},
				Type: event.ImportSomething,
			})
			line = string(byteArrayLine)
		}

		processedLines <- line // Send processed line to intermediate channel to be read by View()
	}
}
