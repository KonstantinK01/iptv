package postgres

import (
	"context"
	"fmt"
)

var seedUsers = []struct {
	login, password string
	active          bool
}{
	{"ivan", "ivan123", true},
	{"anna", "anna123", true},
	{"dmitry", "dmitry123", false},
}

var seedChannels = []struct {
	number int
	name   string
	slug   string
	active bool
}{
	{1, "Первый", "ch1", true},
	{2, "Россия 1", "ch2", true},
	{3, "Матч ТВ", "ch3", true},
	{4, "НТВ", "ch4", true},
	{5, "ТНТ", "ch5", true},
	{6, "Архивный канал", "ch6", false},
}

func (s *Store) Seed(ctx context.Context, hash func(string) (string, error)) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("seed: begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var n int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
		return fmt.Errorf("seed: count users: %w", err)
	}
	if n == 0 {
		for _, u := range seedUsers {
			h, err := hash(u.password)
			if err != nil {
				return fmt.Errorf("seed: hash password: %w", err)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO users (login, password_hash, active) VALUES ($1, $2, $3)
				 ON CONFLICT (login) DO NOTHING`, u.login, h, u.active); err != nil {
				return fmt.Errorf("seed: insert user: %w", err)
			}
		}
	}

	if err := tx.QueryRow(ctx, `SELECT count(*) FROM channels`).Scan(&n); err != nil {
		return fmt.Errorf("seed: count channels: %w", err)
	}
	if n == 0 {
		for _, c := range seedChannels {
			if _, err := tx.Exec(ctx,
				`INSERT INTO channels (number, name, stream_slug, active) VALUES ($1, $2, $3, $4)
				 ON CONFLICT (number) DO NOTHING`, c.number, c.name, c.slug, c.active); err != nil {
				return fmt.Errorf("seed: insert channel: %w", err)
			}
		}
	}
	return tx.Commit(ctx)
}
