package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rokuosan/al/cmd"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	rootCmd := cmd.NewRootCmd(stdout, stderr)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
