package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	op := flag.String("op", "sum", "operation to perform")
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
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOperation, op)
	}

	consolidate := make([]float64, 0)
	for _, file := range filenames {
		// open file for reading
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("Cannot open file: %w", err)
		}
		// parse the csv data
		data, err := csv2float(f, col)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}

		consolidate = append(consolidate, data...)
	}

	_, err := fmt.Fprintln(out, opFunc(consolidate))
	if err != nil {
		return err
	}

	return nil
}
