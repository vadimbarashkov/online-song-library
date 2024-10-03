package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/vadimbarashkov/online-song-library/internal/app"
	"github.com/vadimbarashkov/online-song-library/internal/config"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
