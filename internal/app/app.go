package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/bcrypt"

	"github.com/KonstantinK01/iptv/internal/adapter/postgres"
	"github.com/KonstantinK01/iptv/internal/config"
)

func Run(log *slog.Logger) error {
	c := config.New()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := postgres.Open(ctx, c.DatabaseURL)
	if err != nil {
		return err
	}
	defer st.Close()

	if err := st.Migrate(ctx); err != nil {
		return err
	}
	if err := st.Seed(ctx, hashPassword); err != nil {
		return err
	}

	return nil
}

func hashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}
