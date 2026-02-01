package main

import (
	"os"

	"github.com/ohhmaar/docmd/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
