package event

import (
	"github.com/vbauerster/mpb/v8"
	"github.com/wyattfry/tfinline/inline"
	"testing"
)

func TestFindAddress(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{
			name: "Diagnostic with address",
			event: Event{
				Diagnostic: &Diagnostic{Address: "test_address"},
			},
			expected: "test_address",
		},
		{
			name: "Hook with resource address",
			event: Event{
				Hook: map[string]interface{}{
					"resource": map[string]interface{}{"addr": "hook_address"},
				},
			},
			expected: "hook_address",
		},
		{
			name:     "Default to provider",
			event:    Event{},
			expected: "provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.event.FindAddress()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHandle(t *testing.T) {
	lines := map[string]*inline.Line{
		"test_address": inline.NewLine(mpb.New(), "test_address", "Initial message"),
	}

	tests := []struct {
		name     string
		event    Event
		address  string
		msg      string
		expected inline.Status
		skip     bool
	}{
		{
			name:     "Apply Start",
			event:    Event{Type: EventApplyStart},
			address:  "test_address",
			msg:      "Applying",
			expected: inline.StatusInProgress,
			skip:     false,
		},
		{
			name:     "Apply Complete",
			event:    Event{Type: EventApplyComplete},
			address:  "test_address",
			msg:      "Applied",
			expected: inline.StatusDone,
			skip:     false,
		},
		{
			name: "Apply Errored - Already Exists",
			event: Event{
				Type:    EventApplyErrored,
				Message: "Resource already exists",
			},
			address:  "test_address",
			msg:      "Error",
			expected: inline.StatusFailed,
			skip:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := lines[tt.address]
			_, skip := tt.event.Handle(tt.address, tt.msg, lines)

			if line.Status() != tt.expected {
				t.Errorf("expected status %v, got %v", tt.expected, line.Status())
			}

			if skip != tt.skip {
				t.Errorf("expected skip %v, got %v", tt.skip, skip)
			}
		})
	}
}
