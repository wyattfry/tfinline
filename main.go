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

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

/* ───────────────────────────────────────────── data model ─────────────── */

type Event struct {
	Type string         `json:"type"`
	Hook map[string]any `json:"hook"`
	Msg  string         `json:"@message"`
}

/* ───────────────────────────────────────────── entry point ────────────── */

func main() {
	args := os.Args[1:]
	pretty := isPrettyCmd(args)

	cmd := buildCmd(args, pretty)
	stdout, err := cmd.StdoutPipe()
	must(err)
	cmd.Stderr = os.Stderr
	must(cmd.Start())

	if pretty {
		rows := runPrettyUI(stdout) // returns when TF ends
		fmt.Println()               // blank line after TUI exit
		for _, row := range rows {  // print snapshot
			fmt.Println(row)
		}
	} else {
		passThrough(stdout)
	}

	must(cmd.Wait())
}

/* ───────────────────────────────────── helpers / plumbing ─────────────── */

func isPrettyCmd(args []string) bool {
	return len(args) > 0 && slices.Contains([]string{"apply", "destroy"}, args[0])
}

func buildCmd(args []string, pretty bool) *exec.Cmd {
	if pretty {
		args = append(args, "-auto-approve", "-json")
	}
	return exec.Command("terraform", args...)
}

func passThrough(rdr io.Reader) {
	sc := bufio.NewScanner(rdr)
	for sc.Scan() {
		fmt.Println(sc.Text())
	}
}

/* ────────────────────────────── pretty‑mode TUI renderer ──────────────── */

func runPrettyUI(rdr io.Reader) []string {
	app := tview.NewApplication()
	table := tview.NewTable().SetBorders(false).SetFixed(1, 0)

	bold := tcell.AttrBold
	table.SetCell(0, 0, tview.NewTableCell("RESOURCE").SetAttributes(bold))
	table.SetCell(0, 1, tview.NewTableCell("STATUS").SetAttributes(bold))
	app.SetRoot(table, true)

	scDone := make(chan struct{})

	go func() {
		sc := bufio.NewScanner(rdr)
		for sc.Scan() {
			var ev Event
			if json.Unmarshal(sc.Bytes(), &ev) != nil || ev.Hook["resource"] == nil {
				continue
			}

			addr, _ := ev.Hook["resource"].(map[string]any)["addr"].(string)
			if addr == "" {
				continue
			}

			status := strings.TrimPrefix(ev.Msg, addr+": ")
			row := findOrAddRow(table, addr)

			// colour on completion
			if ev.Type == "apply_complete" || ev.Type == "destroy_complete" {
				status = fmt.Sprintf("[green]%s[-:]", tview.Escape(status))
			}

			table.SetCell(row, 1, tview.NewTableCell(status))
			app.Draw()
		}
		close(scDone)
	}()

	// Run blocks until app.Stop() is called (inside goroutine when stdin closes)
	go func() {
		<-scDone
		app.Stop()
	}()

	_ = app.Run() // alternate screen active during this call

	/* -------- collect snapshot after alternate screen restored -------- */
	var rows []string
	for r := 0; r < table.GetRowCount(); r++ {
		res := table.GetCell(r, 0).Text
		stat := stripTags(table.GetCell(r, 1).Text) // <── CHANGE
		rows = append(rows, fmt.Sprintf("%-50s %s", res, stat))
	}
	return rows
}

/* ─────── tag‑stripping helper ─────── */
var tagRe = regexp.MustCompile(`\[[^][]*]`)

// stripTags removes tview colour / formatting tags like "[green]…[-:]".
func stripTags(s string) string { return tagRe.ReplaceAllString(s, "") }

func findOrAddRow(t *tview.Table, addr string) int {
	for r := 1; r < t.GetRowCount(); r++ {
		if t.GetCell(r, 0).Text == addr {
			return r
		}
	}
	row := t.GetRowCount()
	t.SetCell(row, 0, tview.NewTableCell(addr))
	return row
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
