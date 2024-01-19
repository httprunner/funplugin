//go:build darwin || linux

package myexec

import "testing"

func TestRunShellUnix(t *testing.T) {
	testData := []struct {
		shell          string
		expectExitCode int
	}{
		{"echo hello world", 0},
		{"A=123; echo $A", 0},
		{"A=123 && echo $A", 0},
		{"export A=123 && echo $A", 0},
		{"for i in {1..3}; do echo $i; sleep 1; done", 0},

		{"ls -l; exit 3", 3},
	}

	for _, td := range testData {
		exitCode, err := RunShell(td.shell)
		if exitCode != td.expectExitCode {
			t.Fatalf("expected exit code 0, got %d", exitCode)
		}
		if td.expectExitCode == 0 && err != nil ||
			td.expectExitCode != 0 && err == nil {
			t.Fatal(err)
		}
	}
}
