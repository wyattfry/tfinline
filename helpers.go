package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func isApplyOrDestroy(a []string) bool {
	return len(a) > 0 && slices.Contains([]string{"apply", "destroy"}, a[0])
}

func buildCmd(args []string, pretty bool) *exec.Cmd {
	if pretty {
		args = append(args, "-auto-approve", "-json")
	}
	return exec.Command("terraform", args...)
}

func passThrough(r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		log.Println(sc.Text())
		fmt.Println(sc.Text())
	}
}

func must(err error) {
	if err != nil {
		fmt.Println("Encountered an error, stopping.", err)
		log.Println(err)
		panic(err)
	}
}

func runPretty(r io.Reader) []Event {
	p := mpb.New(mpb.WithWidth(72), mpb.WithRefreshRate(120*time.Millisecond))
	alreadyExistsEvents := []Event{}

	type resInfo struct {
		bar    *mpb.Bar
		status *string // pointer so decorator sees live updates
	}
	bars := map[string]*resInfo{}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		log.Println(sc.Text())
		var ev Event
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		log.Printf("HANDLING NEW Event: %+v\n", ev)
		if ev.Type == "change_summary" && ev.Changes != nil && ev.Changes.Operation != "plan" {
			log.Println("Received change summary and op wasn't plan, marking all bars as done")
			for _, bar := range bars {
				bar.bar.SetCurrent(1) // stop spinner, mark done
			}
		}
		// ignore warnings
		if ev.Type == "diagnostic" && ev.Diagnostic != nil &&
			ev.Diagnostic.Severity == "warning" {
			continue
		}
		address := ""
		if hook, ok := ev.Hook["resource"].(map[string]interface{}); ok {
			if addr := hook["addr"].(string); addr != "" {
				address = addr
			}
		}
		if address == "" && ev.Diagnostic != nil {
			address = ev.Diagnostic.Address
		}

		if address == "" {
			log.Println("No address found in event: ", ev)
			continue
		} else {
			log.Println("Address found in event:", address)
		}

		if !slices.Contains([]string{"apply_start", "apply_complete", "apply_errored", "diagnostic"}, ev.Type) {
			log.Println("We are not interested in this event type:", ev.Type)
			continue
		}

		msg := trimAddrPrefix(ev.Message, address)
		info, seen := bars[address]
		if !seen {
			// create spinner bar with a dynamic “status” decorator
			status := msg
			info = &resInfo{status: &status}

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
		log.Println("CEHCKING IF ALREADY EXISTS", address, ev.Message)
		if strings.Contains(ev.Message, "already exists") {
			log.Println("EXISTS. QUEUE FOR IMPORT", address)
			// If the resource already exists, we can mark it as done
			alreadyExistsEvents = append(alreadyExistsEvents, ev)
			*info.status = "Already exists, queueing for import."
		} else if strings.Contains(ev.Message, "Missing required argument") {
			*info.status = fmt.Sprintf("%s: %s", ev.Message, ev.Diagnostic.Detail)
		} else {
			log.Println("DOES NOT EXISTS. JUST UPDATING STATUS", address)
			*info.status = msg // live update decorator text
		}

		log.Println("SEEING IF MSG IS DONE:", msg)
		if done(msg) {
			log.Println("Marking as done:", address)
			info.bar.SetCurrent(1) // stop spinner, mark done
			gray := "\033[37m"
			reset := "\033[0m"
			*info.status = fmt.Sprintf("%s%s%s", gray, *info.status, reset)
		} else {
			log.Println("NOT DONE", msg)
		}
	}

	p.Wait()

	return alreadyExistsEvents
}

func done(s string) bool {
	ls := strings.ToLower(s)
	return strings.Contains(ls, "complete after") ||
		strings.Contains(ls, "errored after") ||
		strings.Contains(ls, "already exists") ||
		strings.Contains(ls, "missing required argument") ||
		strings.Contains(ls, "failed")
}

func trimAddrPrefix(msg, addr string) string {

	// remove the resource ID from the message, they're long.
	msg = regexp.MustCompile(` \[id=.*\]`).ReplaceAllString(msg, "")

	if strings.HasPrefix(msg, addr+": ") {
		return msg[len(addr)+2:]
	}
	return msg
}
