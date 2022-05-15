package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	DLayout = "2006-01-02"
)

func filterOut(path, ext, mDate string, minSize uint64, info os.FileInfo) bool {
	if info.IsDir() || info.Size() < int64(minSize) {
		return true
	}

	if mDate != "" {
		d, err := time.Parse(DLayout,mDate);
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if info.ModTime().Before(d) {
			return true
		}
	}

	exts := strings.Split(ext, ",")
	for _, ext := range exts {
		if filepath.Ext(path) == ext {
			 return true
		}
	}

	return false
}

func listFile(path string, out io.Writer) error {
	_, err := fmt.Fprintln(out, path)
	return err
}

func delFile(path string, delLogger *log.Logger) error {
	if err := os.Remove(path); err != nil {
		return err
	}

	delLogger.Println(path)  // print to the logger
	return nil
}

func archiveFile(destDir, root, path string) error {
	info, err := os.Stat(destDir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", destDir)
	}

	relDir, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return err
	}

	dest := fmt.Sprintf("%s.gz", filepath.Base(path))
	targetPath := filepath.Join(destDir, relDir, dest)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	// zw := gzip.NewWriter(out)
	zw, err := gzip.NewWriterLevel(out, 9)
	if err != nil {
		return err
	}

	if _, err := io.Copy(zw, in); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return err
	}

	return nil
}

func unarchiveFile(destDir, root, path string) error {
	// do the checks
	info, err := os.Stat(destDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", destDir)
	}
	relDir, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return err
	}
	if filepath.Ext(path) != ".gz" {
		return fmt.Errorf("%q is not a archive file", path)
	}

	// prepare paths
	dest := filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
	targetPath := filepath.Join(destDir, relDir, dest)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	// open files
	out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	// start decompressing
	zR, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, zR); err != nil {
		return err
	}
	return zR.Close()
}
