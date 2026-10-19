package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KonstantinK01/iptv/internal/controller"
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

	handler, err := controller.NewHandler(controller.Deps{
		Users:         st,
		Channels:      st,
		StreamBaseURL: c.StreamBaseURL,
		Log:           log,
	})
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              c.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	log.Info("listening", "addr", c.ListenAddr)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func hashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}
