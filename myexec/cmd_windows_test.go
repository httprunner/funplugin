//go:build windows

package myexec

import "testing"

func TestRunShellWindows(t *testing.T) {
	exitCode, err := RunShell("echo hello world")
	if err != nil {
		t.Fatal(err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	exitCode, err = RunShell("dir /b /s; exit /b 3")
	if err == nil {
		t.Fatal(err)
	}
	t.Log(exitCode)
}
