package main

import (
	"os"
	"testing"
)

func TestFilterOut(t *testing.T) {
	testCases := []struct {
		name string
		file string
		ext string
		minSize uint64
		date string
		expected bool
	} {
		{"FilterNoExtension", "testdata/dir.log", "", 0, "1901-01-01", false},
		{"FilterExtensionMatch", "testdata/dir.log", ".log", 0, "1901-01-01", false},
		{"FilterExtensionNoMatch", "testdata/dir.log", ".sh", 0, "1901-01-01", true},
		{"FilterExtensionSizeMatch", "testdata/dir.log", ".log", 10, "1901-01-01", false},
		{"FilterExtensionSizeNoMatch", "testdata/dir.log", ".log", 20, "1901-01-01", true},
		{"FilterExtensionDateMatch", "testdata/dir.log", ".log", 0, "2022-05-01", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, err := os.Stat(tc.file)
			if err != nil {
				t.Fatal(err)
			}

			f := filterOut(tc.file, tc.ext, tc.date, tc.minSize, info)

			if f != tc.expected {
				t.Errorf("Expected '%t', got '%t' instead\n", tc.expected, f)
			}
		})
	}
}
