package inline

import (
	"testing"

	"github.com/vbauerster/mpb/v8"
)

func TestLineLifecycle(t *testing.T) {
	prog := mpb.New()
	address := "test_resource"
	initialMsg := "Initializing"
	line := NewLine(prog, address, initialMsg)

	if line.Status() != StatusNone {
		t.Errorf("expected status %v, got %v", StatusNone, line.Status())
	}

	line.MarkAsInProgress("In progress")
	if line.Status() != StatusInProgress {
		t.Errorf("expected status %v, got %v", StatusInProgress, line.Status())
	}

	line.MarkAsFailed("Failed")
	if line.Status() != StatusFailed {
		t.Errorf("expected status %v, got %v", StatusFailed, line.Status())
	}

	line.MarkAsDone("Done")
	if line.Status() != StatusDone {
		t.Errorf("expected status %v, got %v", StatusDone, line.Status())
	}
}

func TestSetMessage(t *testing.T) {
	prog := mpb.New()
	line := NewLine(prog, "test_resource", "Initial message")

	newMsg := "Updated message"
	line.SetMessage(newMsg)

	if *line.message != newMsg {
		t.Errorf("expected message %q, got %q", newMsg, *line.message)
	}
}
