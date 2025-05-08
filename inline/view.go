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
		log.Printf("HANDLING EVENT TYPE '%s'\tRESOURCE '%s'", ev.Type, ev.GetAddress())
		log.Printf(rawEventString)

		if ev.Level == "" {
			continue
		}

		if ev.Type == event.Version {
			fmt.Println(ev.Message)
			continue
		}

		if ev.Type == event.InitOutput {
			fmt.Println(ev.Message)
			continue
		}

		if ev.Type == event.TypeChangeSummary {
			if ev.Changes != nil {
				switch ev.Changes.Operation {
				case "apply", "destroy":
					errors += ev.Message + "\n"
				case "plan":
					fmt.Println(ev.Message)
				}
			}
			continue
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
