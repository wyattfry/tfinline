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
		panic(err)
	}
}

func runPretty(r io.Reader) {
	p := mpb.New(mpb.WithWidth(72), mpb.WithRefreshRate(120*time.Millisecond))

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
		if ev.Type == "change_summary" && ev.Changes != nil && ev.Changes.Operation != "plan" {
			for _, bar := range bars {
				bar.bar.SetCurrent(1) // stop spinner, mark done
			}
		}
		// ignore warnings
		if ev.Type == "diagnostic" && ev.Diagnostic != nil &&
			ev.Diagnostic.Severity == "warning" {
			continue
		}
		hook, ok := ev.Hook["resource"].(map[string]interface{})
		if !ok {
			continue
		}
		addr := hook["addr"].(string)
		if addr == "" {
			continue
		}

		msg := trimAddrPrefix(ev.Message, addr)
		info, seen := bars[addr]
		if !seen {
			// create spinner bar with a dynamic “status” decorator
			status := msg
			info = &resInfo{status: &status}

			info.bar = p.New(1, mpb.SpinnerStyle(),
				mpb.PrependDecorators(
					decor.Name(addr, decor.WCSyncSpaceR),
				),
				mpb.AppendDecorators(
					decor.Any(func(_ decor.Statistics) string { return *info.status }),
				),
				mpb.BarWidth(1),
				mpb.BarFillerOnComplete("✅"),
			)

			bars[addr] = info
		}
		*info.status = msg // live update decorator text

		if done(msg) {
			info.bar.SetCurrent(1) // stop spinner, mark done
		}
	}

	p.Wait()
}

func done(s string) bool {
	ls := strings.ToLower(s)
	return strings.Contains(ls, "complete after") ||
		strings.Contains(ls, "errored after") ||
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
