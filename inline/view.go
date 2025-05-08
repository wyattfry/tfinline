package inline

import (
	"encoding/json"
	"fmt"
	"github.com/vbauerster/mpb/v8"
	"github.com/wyattfry/tfinline/event"
	"github.com/wyattfry/tfinline/util"
	"log"
	"time"
)

func View(in <-chan string, done chan<- struct{}) {
	pool := mpb.New(mpb.WithWidth(72), mpb.WithRefreshRate(120*time.Millisecond))
	bars := map[string]*Line{}

	errors := ""

	for rawEventString := range in {
		ev := unmarshalEvent(rawEventString)
		log.Printf("HANDLING EVENT TYPE '%s'\tRESOURCE '%s'\tMESSAGE: '%+v'", ev.Type, ev.GetAddress(), ev.Message)

		if ev.Level == "" {
			continue
		}
		if ev.Type == event.TypeChangeSummary && ev.Changes != nil && ev.Changes.Operation != "refresh" && ev.Changes.Operation != "plan" {
			errors += ev.Message + "\n"
		}
		if ev.Type == event.RefreshStart || ev.Type == event.RefreshComplete || ev.Type == event.RefreshErrored {
			continue
		}
		address := ev.GetAddress()
		if address == "" {
			continue
		}
		bar, exists := bars[address]
		msg := util.TrimAddrPrefix(ev.Message, address)

		if !exists {
			bars[address] = NewLine(pool, address, msg)
			bars[address].MarkAsInProgress(msg)
		}

		switch ev.Type {
		case event.TypeDiagnostic:
			errors += ev.Message + "\n"

		case event.ApplyProgress, event.ApplyStart:
			if bar != nil && bar.Status() == StatusDone {
				bars[address] = NewLine(pool, address, msg)
			}
			if bar != nil {
				bar.MarkAsInProgress(msg)
			}
		case event.ApplyErrored:
			bar.MarkAsFailed(msg)
		case event.ApplyComplete:
			bar.MarkAsDone(msg)
		}
	}

	pool.Wait() // flush progress UI
	if errors != "" {
		fmt.Println(errors)
	}
	done <- struct{}{}
}

func unmarshalEvent(l string) (e *event.Event) {
	var ev event.Event
	if json.Unmarshal([]byte(l), &ev) != nil {
		return &event.Event{
			Message: l,
		}
	}
	return &ev
}
