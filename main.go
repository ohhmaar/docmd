package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/ohhmaar/docmd/cmd"
)

func main() {
	_ = godotenv.Load()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
