package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name string
		root string
		cfg config
		expected string
	} {
		{name: "NoFilter", root: "testdata",
		cfg: config{ext: "", size: 0, list: true},
		expected: "testdata/dir.log\ntestdata/dir2/script.sh\n"},

		{name: "FilterExtensionMatch", root: "testdata",
		cfg: config{ext: ".log", size: 0, list: true},
		expected: "testdata/dir.log\n"},

		{name: "FilterExtensionSizeMatch", root: "testdata",
		cfg: config{ext: ".log", size: 10, list: true},
		expected: "testdata/dir.log\n"},

		{name: "FilterExtensionSizeNoMatch", root: "testdata",
		cfg: config{ext: ".log", size: 20, list: true},
		expected: ""},

		{name: "FilterExtensionNoMatch", root: "testdata",
		cfg: config{ext: ".gz", size: 0, list: true},
		expected: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing. T) {
			var buf bytes.Buffer

			if err := run(tc.root, &buf, tc.cfg); err != nil {
				t.Fatal(err)
			}

			res := buf.String()
			if res != tc.expected {
				t.Errorf("Expected %q, got %q instead\n", tc.expected, res)
			}
		})
	}
}

func createTempDir(t *testing.T, files map[string]int) (dirname string, cleanup func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "walktest")
	if err != nil {
		t.Fatal(err)
	}

	for k, n := range files {
		for j := 0; j < n; j++ {
			fname := fmt.Sprintf("file%d%s", j, k)
			fpath := filepath.Join(tempDir, fname)
			if err := os.WriteFile(fpath, []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	return tempDir, func() {
		os.RemoveAll(tempDir)
	}
}

func TestRunDelExtenstion(t *testing.T) {
	testCases := []struct {
		name string
		cfg config
		extNoDelete string
		nDelete int
		nNoDelete int
		expected string
	} {
		{
			name: "DeleteExtensionNoMatch",
			cfg: config{ext: ".log", del: true},
			extNoDelete: ".gz",
			nDelete: 0,
			nNoDelete: 10,
			expected: "",
		},
		{
			name: "DeleteExtensionatch",
			cfg: config{ext: ".log", del: true},
			extNoDelete: ".gz",
			nDelete: 10,
			nNoDelete: 0,
			expected: "",
		},
		{
			name: "DeleteExtensionMixed",
			cfg: config{ext: ".log", del: true},
			extNoDelete: ".gz",
			nDelete: 5,
			nNoDelete: 5,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				buffer bytes.Buffer
				logBuffer bytes.Buffer
			)

			tc.cfg.wLog = &logBuffer

			tempDir, cleanup := createTempDir(t, map[string]int {
				tc.cfg.ext: tc.nDelete,
				tc.extNoDelete: tc.nNoDelete,
			})

			/* t.Cleanup(func() {  // just like t.Helper, registers function for cleanup
				os.RemoveAll(tempDir)
			}) */
			defer cleanup()

			if err := run(tempDir, &buffer, tc.cfg); err != nil {
				t.Fatal(err)
			}

			res := buffer.String()
			if res != tc.expected {
				t.Errorf("Expected %q, got %q instead\n", tc.expected, res)
			}

			filesLeft, err := os.ReadDir(tempDir)
			if err != nil {
				t.Error(err)
			}
			if len(filesLeft) != tc.nNoDelete {
				for _, n := range filesLeft {
					fmt.Println(n)
				}
				t.Errorf("Expected %d files, got %d instead\n", tc.nNoDelete, len(filesLeft))
			}

			expLogLines := tc.nDelete + 1
			lines := bytes.Split(logBuffer.Bytes(), []byte("\n"))
			if expLogLines != len(lines) {
				t.Errorf("Expected %d log lines, got %d instead", expLogLines, len(lines))
			}
		})
	}
}