package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

func main() {
	op := flag.String("op", "sum", "operation to perform (sum,avg,min,max)")
	column := flag.Int("col", 1, `CSV column to execute operation on
	(starts from 1)`)

	flag.Parse()

	if err := run(flag.Args(), *op, *column, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(filenames []string, op string, col int, out io.Writer) error {
	var opFunc statsFunc

	wg := sync.WaitGroup{}

	if len(filenames) == 0 {
		return ErrNoFiles
	}
	if col < 1 {
		return fmt.Errorf("%w: %d", ErrInvalidColumn, col)
	}

	// validate the column and assign opFunc accordingly
	switch op {
	case "sum":
		opFunc = sum
	case "avg":
		opFunc = avg
	case "min":
		opFunc = min
	case "max":
		opFunc = max
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOperation, op)
	}

	resCh := make(chan []float64)
	errCh := make(chan error)
	doneCh := make(chan struct{})
	filesCh := make(chan string)

	go func() {
		defer close(filesCh)
		for _, file := range filenames {
			filesCh <- file;
		}
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for file := range filesCh {
				// open the file
				f, err := os.Open(file)
				if err != nil {
					errCh <- fmt.Errorf("Cannot open file: %w", err)
					return
				}

				// parse CSV
				res, err := csv2float(f, col)
				if err != nil {
					errCh <- err
				}

				if err := f.Close(); err != nil {
					errCh <- err
				}

				resCh <- res
			}
		}()
	}

	consolidate := make([]float64, 0)

	// wait for all other goroutines to finish (basically wg counter to 0)
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case err := <-errCh:
			return err
		case data := <-resCh:
			consolidate = append(consolidate, data...)
		case <-doneCh:
			_, err := fmt.Fprintln(out, opFunc(consolidate))
			return err
		}
	}
}
