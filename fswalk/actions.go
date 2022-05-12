package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func filterOut(path, ext string, minSize uint64, info os.FileInfo) bool {
	if info.IsDir() || info.Size() < int64(minSize) {
		return true
	}

	if ext != "" && filepath.Ext(path) != ext {
		return true
	}

	return false
}

func listFile(path string, out io.Writer) error {
	_, err := fmt.Fprintln(out, path)
	return err
}

func delFile(path string) error {
	return os.Remove(path)
}
