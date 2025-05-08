package event

type EventType string

const (
	ApplyComplete     EventType = "apply_complete"
	ApplyErrored      EventType = "apply_errored"
	ApplyProgress     EventType = "apply_progress"
	ApplyStart        EventType = "apply_start"
	Outputs           EventType = "outputs"
	PlannedChange     EventType = "planned_change"
	ResourceDrift     EventType = "resource_drift"
	RefreshComplete   EventType = "refresh_complete"
	RefreshErrored    EventType = "refresh_errored"
	RefreshStart      EventType = "refresh_start"
	TypeChangeSummary EventType = "change_summary"
	TypeDiagnostic    EventType = "diagnostic"
	// todo additional?
	// ¯\_(ツ)_/¯
)

type SeverityLevel string

const (
	SeverityLevelWarning SeverityLevel = "warning"
	SeverityLevelError   SeverityLevel = "error"
)
