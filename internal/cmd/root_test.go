package cmd

import (
	"bytes"
	"testing"
)

func executeCommand(args ...string) (stdout, stderr string, err error) {
	root := newRootCmd()
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs(args)
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestRootCommand_Help(t *testing.T) {
	stdout, _, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stdout) == 0 {
		t.Fatal("expected help output, got empty")
	}
}

func TestRootCommand_Version(t *testing.T) {
	stdout, _, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stdout) == 0 {
		t.Fatal("expected version output, got empty")
	}
}
