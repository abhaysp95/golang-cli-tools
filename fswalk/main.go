package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type config struct {
	ext string  // extension to filter out
	size uint64  // minimum file size
	list bool  // list files
	del bool  // delete files
	wLog io.Writer  // to log
	archive string  // archive directory
	mDate string  // specify date
	unarchive bool  // if you want to unarchive
}

func main() {
	root := flag.String("root", ".", "Root directory to start")
	list := flag.Bool("list", false, "List files only")
	ext := flag.String("ext", "", "File extension to filter out (.ext1,.ext2 etc.)")
	size := flag.Uint64("size", 0, "Minimum file size")
	del := flag.Bool("del", false, "Delete files")
	logFile := flag.String("log", "", "Log deletes to the file")
	archive := flag.String("archive", "", "Archive directory")
	date := flag.String("date", "1901-01-01", `List files modified from and onwards the mentioned date
	(format: 2006-02-01)`)
	unarchive := flag.Bool("uarc", false, `set true if want to unarchive,
	set folder from -archive flag`)
	flag.Parse()

	var (
		f = os.Stdout
		err error
	)

	if *logFile != "" {
		f, err = os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	c := config {
		ext: *ext,
		size: *size,
		list: *list,
		del: *del,
		wLog: f,
		archive: *archive,
		mDate: *date,
		unarchive: *unarchive,
	}

	if err := run(*root, os.Stdout, c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, out io.Writer, cfg config) error {
	delLogger := log.New(cfg.wLog, "DELETED FILE: ", log.LstdFlags)

	return filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filterOut(path, cfg.ext, cfg.mDate, cfg.size, info) {
				return nil
			}

			// If list was explicitly set, don't do anything else
			if cfg.list {
				return listFile(path, out)
			}

			if cfg.del {
				return delFile(path, delLogger)
			}

			if cfg.unarchive {
				fmt.Println(cfg.archive)
				if cfg.archive != "" {
					if err := listFile(path, out); err != nil {
						return err
					}
					return unarchiveFile(cfg.archive, root, path)
				}
				return fmt.Errorf("destination not provided")
			}

			if cfg.archive != "" {
				if err := listFile(path, out); err != nil {
					return err
				}
				return archiveFile(cfg.archive, root, path)
			}

			// list is the default option if nothing else was set
			return listFile(path, out)
		})
}

