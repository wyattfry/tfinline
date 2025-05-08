package event

import (
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
			result := tt.event.GetAddress()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
