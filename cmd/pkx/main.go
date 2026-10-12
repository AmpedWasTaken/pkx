package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/AmpedWasTaken/pkx/internal/pkx"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(pkx.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
