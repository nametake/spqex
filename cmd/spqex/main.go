package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nametake/spqex"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] directory\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
}

func main() {
	mode := flag.String("mode", "lint", "Specify mode (lint or fmt). default: lint")
	cmd := flag.String("cmd", "", "Specify command to execute")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("No directory specified.")
		flag.Usage()
		os.Exit(1)
	}
	dir := args[0]

	switch *mode {
	case "fmt":
	case "lint":
	default:
		fmt.Println("Invalid mode specified. Valid modes are fmt or lint.")
		flag.Usage()
		os.Exit(1)
	}

	if *cmd == "" {
		fmt.Println("No command specified.")
		flag.Usage()
		os.Exit(1)
	}

	exitCode, err := run(dir, *cmd, *mode == "fmt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func run(dir string, cmd string, replace bool) (int, error) {
	files, err := spqex.FindGoFiles(dir)
	if err != nil {
		return 0, err
	}

	exitCode := 0
	for i, file := range files {
		result, err := spqex.Process(file, cmd, replace)
		if err != nil {
			return 0, err
		}

		code := result.ExitCode()
		if code != 0 {
			if i != 0 {
				fmt.Fprint(os.Stderr, "\n")
			}
			fmt.Fprintf(os.Stderr, "%s\n", result)
		}

		if code > exitCode {
			exitCode = code
		}

		if result.IsChanged {
			if err := os.WriteFile(file, result.Output, 0); err != nil {
				return 0, fmt.Errorf("failed to write file %s: %v", file, err)
			}
		}
	}

	return exitCode, nil
}
