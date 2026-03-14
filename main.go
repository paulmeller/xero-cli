package main

import (
	"os"

	"github.com/paulmeller/xero-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
