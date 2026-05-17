package afk

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func SignalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	drainCh := make(chan struct{})
	ctx = context.WithValue(ctx, drainingKey{}, drainCh)

	sigCh := make(chan os.Signal, 3)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		// First signal: drain
		<-sigCh
		close(drainCh)

		// Second signal: cancel context (triggers SIGINT to agents)
		<-sigCh
		cancel()

		// Third signal: hard exit
		<-sigCh
		os.Exit(1)
	}()

	return ctx
}
