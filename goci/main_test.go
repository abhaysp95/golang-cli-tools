package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	/* _, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git not installed. Test skipped!!!")
	} */

	testCases := []struct {
		name     string
		proj     string
		out      string
		expErr   error
		setupGit bool
		mockCmd func(ctx context.Context, exe string, args ...string) *exec.Cmd
	}{
		{
			name: "success",
			proj: "./testdata/tool",
			out: "Go Build: SUCCESS\n" +
				"Go Test: SUCCESS\n" +
				"Gofmt: SUCCESS\n" +
				"Git Push: SUCCESS\n",
			expErr:   nil,
			setupGit: true,
			mockCmd:  nil,
		},
		{
			name: "successMock",
			proj: "./testdata/tool",
			out: "Go Build: SUCCESS\n" +
				"Go Test: SUCCESS\n" +
				"Gofmt: SUCCESS\n" +
				"Git Push: SUCCESS\n",
			expErr:   nil,
			setupGit: true,
			mockCmd:  mockCmdContext,
		},
		{
			name:     "fail",
			proj:     "./testdata/toolErr",
			out:      "",
			expErr:   &stepErr{step: "go build"},
			setupGit: false,
		},
		{
			name:     "failFormat",
			proj:     "./testdata/toolfmt",
			out:      "",
			expErr:   &stepErr{step: "go fmt"},
			setupGit: false,
		},
		{
			name: "failTimeout",
			proj: "./testdata/tool",
			out: "",
			expErr: context.DeadlineExceeded,
			setupGit: false,
			mockCmd: mockCmdTimeout,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupGit {
				_, err := exec.LookPath("git")
				if err != nil {
					t.Skip("Git not installed. Test Skipped!!!")
				}

				cleanup := setupGit(t, tc.proj)
				defer cleanup()
			}
			if tc.mockCmd != nil {
				command = tc.mockCmd  // overrides command defined in timeoutStep.go
			}

			var out bytes.Buffer

			err := run(tc.proj, &out)

			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected %q, got 'nil' instead", tc.expErr)
				}

				/* if _, ok := tc.expErr.(*stepErr); ok {  // just for testing purpose (safe to remove)
					t.Errorf("type match")
				} */

				if !errors.Is(err, tc.expErr) { // <-- looks like custom Is() is not getting called
					// if errors.Is(errors.Unwrap(err), tc.expErr) {  // this is working
					t.Errorf("Expected error: %q. Got: %q.", tc.expErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected err: %q", err)
			}
			if out.String() != tc.out {
				t.Errorf("Expected output: %q. Got: %q", tc.out, out.String())
			}
		})
	}
}

func setupGit(t *testing.T, proj string) func() { // returns cleanup function
	t.Helper()

	gitExec, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	tempDir, err := os.MkdirTemp("", "gocitest")
	if err != nil {
		t.Fatal(err)
	}
	projPath, err := filepath.Abs(proj)
	if err != nil {
		t.Fatal(err)
	}

	remoteURI := fmt.Sprintf("file://%s", tempDir)
	var gitCmdList = []struct {
		args []string
		dir  string
		env  []string
	}{
		{
			[]string{"init", "--bare"},
			tempDir, nil,
		},
		{
			[]string{"init"},
			projPath, nil,
		},
		{
			[]string{"remote", "add", "origin", remoteURI},
			projPath, nil,
		},
		{
			[]string{"add", "."},
			projPath, nil,
		},
		{
			[]string{"commit", "-m", "test"},
			projPath,
			/* []string{
				"GIT_COMMITER_NAME=test",
				"GIT_COMMITER_EMAIL=test@example.com",
				"GIT_AUTHOR_NAME=test",
				"GIT_COMMITER_EMAIL=test@example.com",
			}, */
			nil,
		},
	}

	for _, g := range gitCmdList {
		gitCmd := exec.Command(gitExec, g.args...)

		gitCmd.Dir = g.dir
		if g.env != nil {
			fmt.Println("1")
			gitCmd.Env = append(gitCmd.Env, g.env...)
		}

		if err := gitCmd.Run(); err != nil {
			fmt.Println("2")
			t.Fatal(err)
		}
	}

	return func() {
		os.RemoveAll(tempDir)
		os.RemoveAll(filepath.Join(projPath, ".git"))
	}
}

// mockCmdContext function mocks the exec.CommandContext() function (it has same signature)
func mockCmdContext(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess"}

	cs = append(cs, exe)
	cs = append(cs, args...)

	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// mockCmdTimeout function simulates a command that times out
func mockCmdTimeout(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cmd := mockCmdContext(ctx, exe, args...)
	cmd.Env = append(cmd.Env, "GO_HELPER_TIMEOUT=1")
	return cmd
}

// TestHelperProcess is function which tries to simulate the command you would
// have wanted to be executed by real function
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("GO_HELPER_TIMEOUT") == "1" {
		time.Sleep(15 * time.Second)
	}

	// match the name of command you would have wanted to be executed by real
	// function
	if os.Args[2] == "git" {
		fmt.Fprintln(os.Stdout, "Everything up-to-date")
		os.Exit(0)
	}

	os.Exit(1)
}
