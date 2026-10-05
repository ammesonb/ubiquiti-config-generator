// Command ubiquiti-config-generator is the entry point for the GitHub App.
package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Usage = func() {
		// Help output is best effort; there is no useful recovery for a failed write.
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Usage: ubiquiti-config-generator [flags]")
		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "Development scaffold; configuration deployment is not implemented yet.")
		flag.PrintDefaults()
	}
	flag.Parse()
	flag.Usage()
}
