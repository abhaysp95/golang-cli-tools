package main

import (
	"bufio"
	"cli_tools/todo"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var todoFileName = ".todo.json"

func main() {
	// parsing command-line flags
	add := flag.Bool("add", false, "Task to be included in ToDo list")
	list := flag.Bool("list", false, "List all tasks")
	complete := flag.Int("complete", 0, "Item to be completed")
	delete := flag.Int("delete", 0, "Item to delete (doesn't matter if task is completed or not)")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	m := flag.Bool("m", false, "Do multiline input from STDIN")
	u := flag.Bool("u", false, "Show uncomplete tasks only")
	flag.Parse()

	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	l := &todo.List{}

	// use the Get method to read ToDo items from file
	if err := l.Get(todoFileName); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Decide what to do based on the number of arguments
	// provided
	switch {
	// print verbose list
	case *verbose:
		switch {
		case *u:
			fmt.Println(l.UncompleteVerbose())
		default:
			fmt.Println(l.Verbose())
		}
	// For no extra arguments, print the list
	case *list:
		// List current ToDo items
		switch {
		case *u:
			fmt.Println(l.Uncomplete())
		default:
			fmt.Println(l)
		}
	case *complete > 0:
		// complete the given item
		if err := l.Complete(*complete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// save the new list
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *add:
		// when any arguments are provided, they will be used as the new task
		t, err := getTask(os.Stdin, *m, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		l.Add(t)  // add task

		// save the new list
		if err := l.Save(todoFileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *delete > 0:
		if err := l.Delete(*delete); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		// invalid flag provided
		fmt.Fprintln(os.Stderr, "Invalid option")
		os.Exit(1)
	}
}

// getTask function decides where to get the descirption for a new task: arguments or STDIN
func getTask(r io.Reader, m bool, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	output := []string{}

	s := bufio.NewScanner(r)

	for s.Scan() {
		if err := s.Err(); err != nil {
			return strings.Join(output, "\n"), err
		}

		if len(s.Text()) == 0 {
			continue
		}

		output = append(output, s.Text())

		if !m {
			break;
		}
	}

	return strings.Join(output, "\n"), nil
}
