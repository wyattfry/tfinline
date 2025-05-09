package event

type EventType string

const (
	ApplyComplete     EventType = "apply_complete"
	ApplyErrored      EventType = "apply_errored"
	ApplyProgress     EventType = "apply_progress"
	ApplyStart        EventType = "apply_start"
	ImportSomething   EventType = "import_something" // not a real terraform event type but needed for tfinline
	InitOutput        EventType = "init_output"
	Outputs           EventType = "outputs"
	PlannedChange     EventType = "planned_change"
	RefreshComplete   EventType = "refresh_complete"
	RefreshErrored    EventType = "refresh_errored"
	RefreshStart      EventType = "refresh_start"
	ResourceDrift     EventType = "resource_drift"
	TypeChangeSummary EventType = "change_summary"
	TypeDiagnostic    EventType = "diagnostic"
	Version           EventType = "version"
)
