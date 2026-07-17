package cli

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestExecuteProcess(t *testing.T) {
	t.Parallel()

	if os.Getenv("CATSYNC_CLI_TEST_PROCESS") != "1" {
		return
	}

	for index, argument := range os.Args {
		if argument == "--" {
			os.Args = append([]string{"CatSync"}, os.Args[index+1:]...)

			Execute()

			return
		}
	}

	os.Exit(99)
}

func TestExecute_ShowsHelp(t *testing.T) {
	t.Parallel()

	for _, flag := range []string{"-h", "--help"} {
		t.Run(flag, func(t *testing.T) {
			t.Parallel()

			// Given: a help flag.
			result := runExecuteProcess(t, flag)

			// When: CatSync parses it.
			// Then: the process exits successfully and prints its flag usage.
			if result.exitCode != 0 {
				t.Fatalf("exit code = %d, want 0; output: %s", result.exitCode, result.output)
			}

			if !strings.Contains(result.output, "Usage of CatSync:") {
				t.Fatalf("help output = %q, want usage", result.output)
			}
		})
	}
}

func TestExecute_ShowsVersionForBothVersionFlags(t *testing.T) {
	t.Parallel()

	for _, flag := range []string{"-version", "-v"} {
		t.Run(flag, func(t *testing.T) {
			t.Parallel()

			// Given: a version flag.
			result := runExecuteProcess(t, flag)

			// When: CatSync parses it.
			// Then: the process exits successfully and prints version metadata.
			if result.exitCode != 0 {
				t.Fatalf("exit code = %d, want 0; output: %s", result.exitCode, result.output)
			}

			if !strings.Contains(result.output, "CatSync[") {
				t.Fatalf("version output = %q, want CatSync version", result.output)
			}
		})
	}
}

func TestExecute_RejectsUnknownFlag(t *testing.T) {
	t.Parallel()

	// Given: an unknown flag.
	result := runExecuteProcess(t, "-unknown")

	// When: CatSync parses it.
	// Then: the process exits with the flag parsing failure code.
	if result.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; output: %s", result.exitCode, result.output)
	}
}

type executeProcessResult struct {
	exitCode int
	output   string
}

func runExecuteProcess(t *testing.T, arguments ...string) executeProcessResult {
	t.Helper()

	commandArguments := append([]string{"-test.run=^TestExecuteProcess$", "--"}, arguments...)
	//nolint:gosec // The test intentionally re-executes the current test binary.
	command := exec.Command(os.Args[0], commandArguments...)

	command.Env = append(os.Environ(), "CATSYNC_CLI_TEST_PROCESS=1")

	output, err := command.CombinedOutput()
	if err == nil {
		return executeProcessResult{output: string(output)}
	}

	exitError := &exec.ExitError{}

	ok := errors.As(err, &exitError)
	if !ok {
		t.Fatalf("run CLI test process: %v", err)
	}

	return executeProcessResult{exitCode: exitError.ExitCode(), output: string(output)}
}
