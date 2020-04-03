package main

import (
	"fmt"
	"os"

	"github.com/tomocy/go-memstats"
)

func main() {
	if err := memstats.Run(func() memstats.Window {
		return memstats.NewGrid()
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
