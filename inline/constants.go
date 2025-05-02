package inline

type Status int

const ( // other statuses needed?
	StatusNone Status = iota
	StatusInProgress
	StatusFailed
	StatusDone
)
