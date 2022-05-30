package main

import (
	"bytes"
	"errors"
	"testing"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name   string
		proj   string
		out    string
		expErr error
	}{
		{
			name:   "success",
			proj:   "./testdata/tool",
			out:    "Go Build: SUCCESS\nGo Test: SUCCESS\nGofmt: SUCCESS\n",
			expErr: nil,
		},
		{
			name:   "fail",
			proj:   "./testdata/toolErr",
			out:    "",
			expErr: &stepErr{step: "go build"},
		},
		{
			name: "failFormat",
			proj: "./testdata/toolfmt",
			out: "",
			expErr: &stepErr{step: "go fmt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer

			err := run(tc.proj, &out)

			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected %q, got 'nil' instead", tc.expErr)
				}

				/* if _, ok := tc.expErr.(*stepErr); ok {  // just for testing purpose (safe to remove)
					t.Errorf("type match")
				} */

				if !errors.Is(err, tc.expErr) {  // <-- looks like custom Is() is not getting called
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
