package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type config struct {
	ext string  // extension to filter out
	size uint64  // minimum file size
	list bool  // list files
	del bool  // delete files
}

func main() {
	root := flag.String("root", ".", "Root directory to start")
	list := flag.Bool("list", false, "List files only")
	ext := flag.String("ext", "", "File extension to filter out")
	size := flag.Uint64("size", 0, "Minimum file size")
	delete := flag.Bool("del", false, "Delete files")
	flag.Parse()

	c := config {
		ext: *ext,
		size: *size,
		list: *list,
		del: *delete,
	}

	if err := run(*root, os.Stdout, c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, out io.Writer, cfg config) error {
	return filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filterOut(path, cfg.ext, cfg.size, info) {
				return nil
			}

			// If list was explicitly set, don't do anything else
			if cfg.list {
				return listFile(path, out)
			}

			if cfg.del {
				/* if err := listFile(path, os.Stdout); err != nil {
					return err
				} */
				return delFile(path)
			}

			// list is the default option if nothing else was set
			return listFile(path, out)
		})
}

