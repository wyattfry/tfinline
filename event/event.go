package event

import "encoding/json"

type ChangeSummary struct {
	Add       int    `json:"add"`
	Change    int    `json:"change"`
	Import    int    `json:"import"`
	Remove    int    `json:"remove"`
	Operation string `json:"operation"`
}

type Diagnostic struct {
	Address string `json:"address"`
	Detail  string `json:"detail"`
}

type Event struct {
	Level      string                 `json:"@level"`
	Type       EventType              `json:"type"`
	Hook       map[string]interface{} `json:"hook"`
	Message    string                 `json:"@message"`
	Diagnostic *Diagnostic            `json:"diagnostic,omitempty"`
	Changes    *ChangeSummary         `json:"changes,omitempty"`
}

func (e Event) GetAddress() string {
	if e.Diagnostic != nil && e.Diagnostic.Address != "" {
		return e.Diagnostic.Address
	}

	if h, ok := e.Hook["resource"].(map[string]interface{}); ok {
		if a := h["addr"].(string); a != "" {
			return a
		}
	}

	return ""
}

func UnmarshalEvent(l string) (e *Event) {
	var ev Event
	if json.Unmarshal([]byte(l), &ev) != nil {
		return &Event{
			Message: l,
		}
	}
	return &ev
}
