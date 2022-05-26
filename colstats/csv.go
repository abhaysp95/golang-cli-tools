package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
)

func sum(data []float64) float64 {
	sum := 0.0
	for _, d := range data {
		sum += d
	}
	return sum
}

func avg(data []float64) float64 {
	return sum(data) / float64(len(data))
}

func min(data []float64) float64 {
	res := math.Inf(1)
	for _, d := range data {
		if res > d {
			res = d
		}
	}

	return res
}

func max(data []float64) float64 {
	res := math.Inf(-1)
	for _, d := range data {
		if res < d {
			res = d
		}
	}

	return res
}

type statsFunc func(data []float64) float64

func csv2float(r io.Reader, column int) ([]float64, error) {
	cr := csv.NewReader(r)
	cr.ReuseRecord = true  // use the same backing array for slice from previous call of read
	column--  // adjust base index to 0

	// real all csv data in one go
	/* allData, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("Can't read data from file: %w", err)
	} */

	var data []float64
	for i := 0; ; i++ {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Can't read data from file: %w", err)
		}
		if i == 0 {
			continue
		}
		if len(row) <= column {
			return nil, fmt.Errorf("%w: File has only %d columns", ErrInvalidColumn, len(row))
		}
		v, err := strconv.ParseFloat(row[column], 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrNotNumber, err)
		}
		data = append(data, v)
	}

	return data, nil
}
