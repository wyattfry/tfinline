package util

import (
	"testing"
)

func TestTrimAddrPrefix(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		addr     string
		expected string
	}{
		{
			name:     "Remove ID and prefix",
			msg:      "azurerm_resource_group.example: Still updating [id=\"a-very-long-resource-id\"]",
			addr:     "azurerm_resource_group.example",
			expected: "Still updating",
		},
		{
			name:     "No ID to remove",
			msg:      "azurerm_resource_group.example: Still updating",
			addr:     "azurerm_resource_group.example",
			expected: "Still updating",
		},
		{
			name:     "No prefix to remove",
			msg:      "Still updating [id=\"a-very-long-resource-id\"]",
			addr:     "azurerm_resource_group.example",
			expected: "Still updating",
		},
		{
			name:     "No changes needed",
			msg:      "Still updating",
			addr:     "azurerm_resource_group.example",
			expected: "Still updating",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimAddrPrefix(tt.msg, tt.addr)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
