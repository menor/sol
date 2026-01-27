package main

import (
	"os"

	"lab.plat.farm/menor/sol/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
