package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/ammesonb/ubiquiti-config-generator/internal/cli"
	"golang.org/x/term"
)

func main() { os.Exit(run()) }

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return cli.Run(ctx, os.Args[1:], cli.Streams{Input: os.Stdin, Output: os.Stdout, Error: os.Stderr, Interactive: term.IsTerminal(int(os.Stdin.Fd()))})
}
