package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetValidationLevel(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		level ValidationLevel
	}{
		{"config", "config", ValidationNone},
		{"config subcommand", "config view", ValidationNone},
		{"version", "version", ValidationNone},
		{"completion", "completion", ValidationNone},
		{"completion bash", "completion bash", ValidationNone},
		{"health", "health", ValidationServer},
		{"login", "login", ValidationServer},
		{"tenant list", "tenant list", ValidationAuth},
		{"tenant select", "tenant select", ValidationAuth},
		{"tenant get", "tenant get", ValidationAuth},
		{"project list", "project list", ValidationFull},
		{"api", "api", ValidationFull},
		{"token list", "token list", ValidationFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.path}
			// Simulate command hierarchy
			parts := splitPath(tt.path)
			if len(parts) > 1 {
				parent := &cobra.Command{Use: parts[0]}
				child := &cobra.Command{Use: parts[1]}
				parent.AddCommand(child)
				cmd = child
			}
			got := getValidationLevel(cmd)
			assert.Equal(t, tt.level, got, "path=%q", tt.path)
		})
	}
}

func splitPath(path string) []string {
	result := []string{}
	current := ""
	for _, c := range path {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
