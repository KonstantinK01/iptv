package main

import (
	"log/slog"
	"os"

	"github.com/KonstantinK01/iptv/internal/app"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := app.Run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}
