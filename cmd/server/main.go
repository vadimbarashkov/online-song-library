package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/vadimbarashkov/online-song-library/internal/app"
	"github.com/vadimbarashkov/online-song-library/internal/config"
)

var configPath = ".env"

func main() {
	if val, ok := os.LookupEnv("CONFIG_PATH"); ok {
		configPath = val
	}

	cfg, err := config.Load(configPath)
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
