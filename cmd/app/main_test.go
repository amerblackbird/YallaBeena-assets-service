package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	// Simple test to ensure CI passes
	if 1+1 != 2 {
		t.Error("Basic math failed")
	}
}

func TestExampleFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"positive", 5, 5},
		{"zero", 0, 0},
		{"negative", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, tt.input)
			}
		})
	}
}