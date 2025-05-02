package event

type EventType string

const (
	EventRefreshStart      EventType = "refresh_start"
	EventRefreshComplete   EventType = "refresh_complete"
	EventRefreshErrored    EventType = "refresh_errored"
	EventApplyStart        EventType = "apply_start"
	EventApplyProgress     EventType = "apply_progress"
	EventApplyComplete     EventType = "apply_complete"
	EventApplyErrored      EventType = "apply_errored"
	EventTypeDiagnostic    EventType = "diagnostic"
	EventTypeChangeSummary EventType = "change_summary"
	// todo additional?
)

type SeverityLevel string

const (
	SeverityLevelWarning SeverityLevel = "warning"
	SeverityLevelError   SeverityLevel = "error"
)
