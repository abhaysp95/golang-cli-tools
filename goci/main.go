package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type executer interface {
	execute() (string, error)
}

func run(pName string, out io.Writer) error {
	if pName == "" {
		return fmt.Errorf("Project directory is required: %w", ErrValidation)
	}

	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error)
	done := make(chan struct{})

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	pipeline := make([]executer, 4)
	pipeline[0] = newStep("go build", "go", "Go Build: SUCCESS",
		pName, []string{"build", ".", "errors"})
	pipeline[1] = newStep("go test", "go", "Go Test: SUCCESS",
		pName, []string{"test", "-v"})
	pipeline[2] = newExceptionStep("go fmt", "gofmt", "Gofmt: SUCCESS",
		pName, []string{"-l", "."})
	pipeline[3] = newTimeoutStep("git push", "git", "Git Push: SUCCESS",
		pName, []string{"push", "origin", "main"}, 15*time.Second)

	go func() {
		for _, s := range pipeline {
			msg, err := s.execute()
			if err != nil {
				errCh <- err
				return
			}

			_, err = fmt.Fprintln(out, msg)
			if err != nil {
				errCh <- err
				return
			}
		}
		close(done)
	}()

	for {
		select {
		case sig := <-sigCh:
			signal.Stop(sigCh)
			return fmt.Errorf("%s: Exiting: %w", sig, ErrSignal)
		case err := <-errCh:
			return err
		case <-done:
			return nil
		}
	}
}

func main() {
	proj := flag.String("p", "", "Project directory")
	flag.Parse()

	if err := run(*proj, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
