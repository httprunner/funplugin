package myexec

import "testing"

func TestRunShell(t *testing.T) {
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
	if exitCode != 3 {
		t.Fatalf("expected exit code 3, got %d", exitCode)
	}
}
