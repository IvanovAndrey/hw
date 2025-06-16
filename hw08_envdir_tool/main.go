package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		if _, err := fmt.Fprintf(os.Stderr, "Usage: %s /path/to/env/dir command [args...]\n", os.Args[0]); err != nil {
			os.Exit(2)
		}
		os.Exit(1)
	}
	dir := os.Args[1]
	cmd := os.Args[2:]
	env, err := ReadDir(dir)
	if err != nil {
		if _, errPrint := fmt.Fprintln(os.Stderr, err); errPrint != nil {
			os.Exit(2)
		}
		os.Exit(1)
	}

	os.Exit(RunCmd(cmd, env))
}
