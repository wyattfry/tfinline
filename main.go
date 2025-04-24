package main

import (
	"fmt"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type Diagnostic struct {
	Severity string `json:"severity"`
	Address  string `json:"address"`
	Detail   string `json:"detail"`
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
	logFile, err := os.OpenFile(".tfinline.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

	var alreadyExistingEvents []Event

	if pretty {
		alreadyExistingEvents = runPretty(stdout)
	} else {
		passThrough(stdout)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("Command finished with error: %v\n", err)
		if len(alreadyExistingEvents) > 0 {
			log.Println("Importing resources that already exist:")
			fmt.Println("### Importing resources that already exist ###")
			p := mpb.New(mpb.WithWidth(72), mpb.WithRefreshRate(120*time.Millisecond))
			type resInfo struct {
				bar    *mpb.Bar
				status *string // pointer so decorator sees live updates
			}
			bars := map[string]*resInfo{}

			// Initialize progress bars for each resource address
			for _, ev := range alreadyExistingEvents {
				address := ev.Diagnostic.Address
				status := "Importing"
				info := &resInfo{status: &status}

				info.bar = p.New(1, mpb.SpinnerStyle(),
					mpb.PrependDecorators(
						decor.Name(address, decor.WCSyncSpaceR),
					),
					mpb.AppendDecorators(
						decor.Any(func(_ decor.Statistics) string { return *info.status }),
					),
					mpb.BarWidth(1),
					mpb.BarFillerOnComplete("✓"),
				)

				bars[address] = info
			}

			// Serially import each resource
			for _, ev := range alreadyExistingEvents {
				address := ev.Diagnostic.Address
				id := regexp.MustCompile(`(ID ")(.*)(" already exists)`).FindStringSubmatch(ev.Message)[2]
				log.Printf("Address: %s, ID: %s\n", address, id)

				// Update status to "Importing"
				info := bars[address]
				*info.status = "Importing"

				// Run the import command
				cmd := exec.Command("terraform", "import", address, id)
				cmd.Stdout = log.Writer()
				cmd.Stderr = log.Writer()
				must(cmd.Start())
				must(cmd.Wait())

				// Mark the progress bar as done
				gray := "\033[37m"
				reset := "\033[0m"
				*info.status = fmt.Sprintf("%s%s%s", gray, "Import Complete.", reset)
				info.bar.SetCurrent(1)
			}

			// Wait for all progress bars to complete
			p.Wait()
		}
	}
}
