package inline

import (
	"fmt"
	"github.com/vbauerster/mpb/v8"
	"github.com/wyattfry/tfinline/event"
	"github.com/wyattfry/tfinline/util"
	"log"
	"strings"
	"time"
)

type handlerInput struct {
	ev                *event.Event
	errors            *string
	progressContainer *mpb.Progress
	bars              map[string]*Line
	toImport          map[string]bool
}

func View(in <-chan string, done chan<- struct{}) {
	progressContainer := mpb.New(mpb.WithWidth(72), mpb.WithRefreshRate(120*time.Millisecond))
	bars := map[string]*Line{}
	toImport := map[string]bool{}
	errors := ""

	handlers := map[string]func(*handlerInput){
		string(event.Version):           handleVersion,
		string(event.InitOutput):        handleInit,
		string(event.TypeChangeSummary): handleChangeSummary,
		string(event.ImportSomething):   handleImport,
		string(event.TypeDiagnostic):    handleDiagnostic,
		string(event.ApplyStart):        handleApplyStart,
		string(event.ApplyProgress):     handleApplyProgress,
		string(event.ApplyErrored):      handleApplyErrored,
		string(event.ApplyComplete):     handleApplyComplete,
	}

	for rawEventString := range in {
		ev := event.UnmarshalEvent(rawEventString)
		log.Printf("HANDLING EVENT TYPE '%s'\tRESOURCE '%s'", ev.Type, ev.GetAddress())
		log.Println(rawEventString)

		if handler, ok := handlers[string(ev.Type)]; ok {
			handler(&handlerInput{
				ev:                ev,
				errors:            &errors,
				progressContainer: progressContainer,
				bars:              bars,
				toImport:          toImport,
			})
		}
	}

	progressContainer.Wait() // flush progressContainer UI

	if errors != "" {
		fmt.Println(errors)
	}

	done <- struct{}{}
}

func handleVersion(input *handlerInput) {
	fmt.Println(input.ev.Message)
}

func handleInit(input *handlerInput) {
	fmt.Println(input.ev.Message)
}

func handleChangeSummary(input *handlerInput) {
	if input.ev.Changes != nil {
		switch input.ev.Changes.Operation {
		case "apply", "destroy":
			*input.errors += input.ev.Message + "\n"
		case "plan":
			fmt.Println(input.ev.Message)
		}
	}
}

func handleImport(input *handlerInput) {
	address := input.ev.GetAddress()
	if strings.Contains(input.ev.Message, "Importing from ID") {
		input.bars[address] = NewLine(input.progressContainer, address, "Importing...")
		input.bars[address].MarkAsInProgress("Importing...")
	}
	if strings.Contains(input.ev.Message, "Import successful") {
		input.bars[address].MarkAsDone("Import Successful")
	}
	if strings.Contains(input.ev.Message, "error") {
		input.bars[address].MarkAsFailed(input.ev.Message)
	}
}

func handleDiagnostic(input *handlerInput) {
	if strings.Contains(input.ev.Message, "Provider development overrides") {
		fmt.Println(input.ev.Message)
		return
	}
	if strings.Contains(input.ev.Message, "error: A resource with the ID") {
		tfobj := util.ExtractResourceAddressAndId(input.ev.Message)
		input.toImport[tfobj.Address] = true
		return
	}
	// The 'captial E' import error doesn't have a resource address, so throw it away
	if strings.Contains(input.ev.Message, "Error: A resource with the ID") {
		return
	}
	*input.errors = fmt.Sprintf("%s\n%s\n", *input.errors, input.ev.Message)
}

func handleApplyStart(input *handlerInput) {
	address := input.ev.GetAddress()
	msg := util.TrimAddrPrefix(input.ev.Message, address)
	input.bars[address] = NewLine(input.progressContainer, address, msg)
}

func handleApplyProgress(input *handlerInput) {
	address := input.ev.GetAddress()
	bar := input.bars[address]
	bar.MarkAsInProgress(util.TrimAddrPrefix(input.ev.Message, address))
}

func handleApplyErrored(input *handlerInput) {
	address := input.ev.GetAddress()
	bar := input.bars[address]
	msg := util.TrimAddrPrefix(input.ev.Message, address)
	bar.MarkAsFailed(msg)
}

func handleApplyComplete(input *handlerInput) {
	address := input.ev.GetAddress()
	bar := input.bars[address]
	msg := util.TrimAddrPrefix(input.ev.Message, address)
	bar.MarkAsDone(msg)
}
