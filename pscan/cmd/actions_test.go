package cmd

import (
	"bytes"
	"cli_tools/pscan/scan"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	// create temp file
	tf, err := os.CreateTemp("", "pScan")
	if err != nil {
		t.Fatal(err)
	}

	if initList {
		hl := &scan.HostsList{}
		for _, host := range hosts {
			if err := hl.Add(host); err != nil {
				t.Fatal(err)
			}
		}
		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestHostActions(t *testing.T) {
	hosts := []string{"host1", "host2", "host3"}
	testCases := []struct {
		name string
		args []string
		expectedOut string
		initList bool
		actionFunc func(io.Writer, string, []string) error
	} {
		{
			name: "AddAction",
			args: hosts,
			expectedOut: "Added host: host1\nAdded host: host2\nAdded host: host3\n",
			initList: false,
			actionFunc: addAction,
		},
		{
			name: "ListAction",
			expectedOut: "host1\nhost2\nhost3\n",
			initList: true,
			actionFunc: listAction,
		},
		{
			name: "DeleteAction",
			args: []string{"host1", "host2"},
			expectedOut: "Deleted host: host1\nDeleted host: host2\n",
			initList: true,
			actionFunc: deleteAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tf, cleanup := setup(t, hosts, tc.initList)
			defer cleanup()

			// capture action output
			var out bytes.Buffer
			if err := tc.actionFunc(&out, tf, tc.args); err != nil {
				t.Fatalf("Expected no error, got %q\n", err)
			}

			if out.String() != tc.expectedOut {
				t.Errorf("Expected output %q, got output %q\n", tc.expectedOut, out.String())
			}
		})

	}
}

func TestIntegration(t *testing.T) {
	hosts := []string{ "host1", "host2", "host3" }
	tf, cleanup := setup(t, hosts, false)
	defer cleanup()

	delHost := "host2"
	endHosts := []string{ "host1", "host3" }

	expectedOut := ""
	for _, h := range hosts {
		expectedOut += fmt.Sprintln("Added host:", h)
	}
	expectedOut += strings.Join(hosts, "\n")
	expectedOut += fmt.Sprintln()
	expectedOut += fmt.Sprintln("Deleted host:", delHost)
	expectedOut += strings.Join(endHosts, "\n")
	expectedOut += fmt.Sprintln()

	var out bytes.Buffer

	// perform all actions in sequence
	if err := addAction(&out, tf, hosts); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	if err := deleteAction(&out, tf, []string{delHost}); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("Expected no error, got %q\n", err)
	}

	// compare with expected output
	outStr := out.String()
	if outStr != expectedOut {
		t.Errorf("Expected output %q, got output %q\n", expectedOut, outStr)
	}
}
