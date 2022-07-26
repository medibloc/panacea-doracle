package main

import (
	"github.com/medibloc/panacea-doracle/cmd/doracled/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
