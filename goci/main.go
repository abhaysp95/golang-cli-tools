package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

type executer interface {
	execute() (string, error)
}

func run(pName string, out io.Writer) error {
	if pName == "" {
		return fmt.Errorf("Project directory is required: %w", ErrValidation)
	}

	pipeline := make([]executer, 4)
	pipeline[0] = newStep("go build", "go", "Go Build: SUCCESS",
		pName, []string{"build", ".", "errors"})
	pipeline[1] = newStep("go test", "go", "Go Test: SUCCESS",
		pName, []string{"test", "-v"})
	pipeline[2] = newExceptionStep("go fmt", "gofmt", "Gofmt: SUCCESS",
		pName, []string{"-l", "."})
	pipeline[3] = newTimeoutStep("git push", "git", "Git Push: SUCCESS",
		pName, []string{"push", "origin", "main"}, 15*time.Second)

	for _, s := range pipeline {
		msg, err := s.execute()
		// fmt.Println("msg: ", msg)
		if err != nil {
			// fmt.Println(err)
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
