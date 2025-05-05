package event

import (
	"github.com/wyattfry/tfinline/inline"
	"log"
	"strings"
)

// "changes":{"add":0,"change":0,"import":0,"remove":3,"operation":"plan"}
type ChangeSummary struct {
	Add       int    `json:"add"`
	Change    int    `json:"change"`
	Import    int    `json:"import"`
	Remove    int    `json:"remove"`
	Operation string `json:"operation"`
}

type Diagnostic struct {
	Severity SeverityLevel `json:"severity"`
	Address  string        `json:"address"`
	Detail   string        `json:"detail"`
}

type Event struct {
	Level      string                 `json:"@level"`
	Type       EventType              `json:"type"`
	Hook       map[string]interface{} `json:"hook"`
	Message    string                 `json:"@message"`
	Diagnostic *Diagnostic            `json:"diagnostic,omitempty"`
	Changes    *ChangeSummary         `json:"changes,omitempty"`
}

func (e Event) FindAddress() string {
	if e.Diagnostic != nil && e.Diagnostic.Address != "" {
		return e.Diagnostic.Address
	}

	if h, ok := e.Hook["resource"].(map[string]interface{}); ok {
		if a := h["addr"].(string); a != "" {
			return a
		}
	}

	return "provider" // get provider dyanmically? e.g azurerm, azuread if run in multiple?
}

func (e Event) Handle(address, msg string, lines map[string]*inline.Line) (*Event, bool) {
	l := lines[address]

	// handle provider (or rename terraform (run)? can output result) address separately somehow
	//

	switch e.Type {
	case EventRefreshStart:
		l.MarkAsInProgress(msg)
	case EventRefreshComplete:
		// should not mark as done if command is not `refresh`
		// if command is refresh, then markasdone
		l.MarkAsInProgress(msg)
	case EventRefreshErrored:
		l.MarkAsFailed(msg)
	case EventApplyStart:
		l.MarkAsInProgress(msg)
	case EventApplyProgress:
		l.MarkAsInProgress(msg)
	case EventApplyComplete:
		log.Println("Marking as done:", address)
		l.MarkAsDone(msg)
	case EventApplyErrored:
		if strings.Contains(e.Message, "already exists") {
			log.Println("EXISTS. QUEUE FOR IMPORT", address)
			l.MarkAsFailed("Already exists, queueing for import.")
			return &e, true
		}

		return nil, false
		// TODO: mark as failed and return / store error?
	case EventTypeDiagnostic:
		if e.Diagnostic == nil {
			// todo what??
		}

		switch e.Diagnostic.Severity {
		case SeverityLevelWarning:
			if address == "provider" {
				l.MarkAsInProgress(msg)
				return nil, false
			}
			return nil, true
		case SeverityLevelError:
			// todo handle
		}
	case EventTypeChangeSummary:
		if e.Changes != nil && e.Changes.Operation != "plan" {
			log.Println("Received change summary and op wasn't plan, marking all bars as done")
			for _, l := range lines {
				l.MarkAsDone("")
			}
		}
	default:
		// default to in progress?
		l.MarkAsInProgress(msg)
	}

	// default to not skipping out of loop to let bars update unless we skip explicitly
	return nil, false
}
