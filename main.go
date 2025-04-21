// tfinline-mpb.go  –  docker‑compose‑style live status for `terraform apply|destroy`
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

/* ─────────────── Terraform JSON event ────────────────── */

type Diagnostic struct {
	Severity string `json:"severity"`
}

type Event struct {
	Type       string                 `json:"type"`
	Hook       map[string]interface{} `json:"hook"`
	Message    string                 `json:"@message"`
	Diagnostic *Diagnostic            `json:"diagnostic,omitempty"`
}

/* ─────────────── main ───────────────────────────────── */

func main() {
	args := os.Args[1:]
	pretty := isApplyOrDestroy(args)

	cmd := buildCmd(args, pretty)
	stdout, err := cmd.StdoutPipe()
	must(err)
	cmd.Stderr = os.Stderr
	must(cmd.Start())

	if pretty {
		runPretty(stdout)
	} else {
		passThrough(stdout)
	}

	must(cmd.Wait())
}

/* ─────────────── helpers & plumbing ─────────────────── */

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
		fmt.Println(sc.Text())
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

/* ───────────────── live status with mpb ───────────────── */

func runPretty(r io.Reader) {
	p := mpb.New(mpb.WithWidth(72), mpb.WithRefreshRate(120*time.Millisecond))

	type resInfo struct {
		bar    *mpb.Bar
		status *string // pointer so decorator sees live updates
	}
	bars := map[string]*resInfo{}

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var ev Event
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
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
				mpb.PrependDecorators(decor.Name(addr+": ", decor.WCSyncWidth)),
				mpb.AppendDecorators(
					decor.Any(func(s decor.Statistics) string { return *info.status }, decor.WCSyncWidth),
				),
				mpb.BarWidth(1),
			)
			bars[addr] = info
		}
		*info.status = msg // live update decorator text

		if done(msg) {
			info.bar.SetCurrent(1) // stop spinner, mark done
		}
	}

	// abort any left hanging (e.g., skipped by TF)
	// for _, i := range bars {
	// 	if i.bar.Completed() {
	// 		i.bar.Abort(true)
	// 	}
	// }
	p.Wait()

	// -------- snapshot rows for scroll‑back --------
	// var rows []string
	// for addr, i := range bars {
	// 	rows = append(rows, fmt.Sprintf("%-55s %s", addr, stripAnsi(*i.status)))
	// }
	// sort.Strings(rows)
	// return rows
}

/* ───────────────── misc string helpers ───────────────── */

func done(s string) bool {
	ls := strings.ToLower(s)
	return strings.Contains(ls, "complete") ||
		strings.Contains(ls, "error") ||
		strings.Contains(ls, "failed")
}

func trimAddrPrefix(msg, addr string) string {
	if strings.HasPrefix(msg, addr+": ") {
		return msg[len(addr)+2:]
	}
	return msg
}

// stripAnsi removes any ANSI colour codes mpb’s spinner may emit
var ansiRe = regexp.MustCompile("\033\\[[0-9;]*m")

func stripAnsi(s string) string { return ansiRe.ReplaceAllString(s, "") }
