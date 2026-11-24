package main

import (
	"os"

	"github.com/pablogtrz/tablon/cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Getenv, os.Stdin, os.Stdout, os.Stderr))
}
