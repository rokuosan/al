package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantOut     string
		wantErr     string
		wantNoError bool
	}{
		{
			name:        "help by default",
			args:        nil,
			wantOut:     "Contextual aliases for your shell",
			wantNoError: true,
		},
		{
			name:    "list exists",
			args:    []string{"list"},
			wantErr: "list is not implemented yet",
		},
		{
			name:    "run requires task name",
			args:    []string{"run"},
			wantErr: "requires at least 1 arg",
		},
		{
			name:    "run exists",
			args:    []string{"run", "hello"},
			wantErr: "run is not implemented yet",
		},
		{
			name:    "doctor exists",
			args:    []string{"doctor"},
			wantErr: "doctor is not implemented yet",
		},
		{
			name:    "init requires shell",
			args:    []string{"init"},
			wantErr: "accepts 1 arg",
		},
		{
			name:    "init exists",
			args:    []string{"init", "zsh"},
			wantErr: "init is not implemented yet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			err := run(tt.args, &stdout, &stderr)
			if tt.wantNoError {
				if err != nil {
					t.Fatalf("run() error = %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
				}
			}

			if tt.wantOut != "" && !strings.Contains(stdout.String(), tt.wantOut) {
				t.Fatalf("stdout = %q, want substring %q", stdout.String(), tt.wantOut)
			}
		})
	}
}
