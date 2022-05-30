package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func run(pName string, out io.Writer) error {
	if pName == "" {
		return fmt.Errorf("Project directory is required: %w", ErrValidation)
	}

	pipeline := make([]step, 2)
	pipeline[0] = newStep("go build", "go", "Go Build: SUCCESS",
		pName, []string{"build", ".", "errors"})
	pipeline[1] = newStep("go test", "go", "Go Test: SUCCESS",
		pName, []string{"test", "-v"})

	for _, s := range pipeline {
		msg, err := s.execute()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(out, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	proj := flag.String("p", "", "Project directory")
	flag.Parse()

	if err := run(*proj, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
