package postgres

import (
	"context"
	"fmt"
)

var schema = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id            BIGSERIAL PRIMARY KEY,
		login         TEXT    NOT NULL UNIQUE,
		password_hash TEXT    NOT NULL,
		active        BOOLEAN NOT NULL DEFAULT TRUE
	)`,
	`CREATE TABLE IF NOT EXISTS channels (
		id          BIGSERIAL PRIMARY KEY,
		number      INTEGER NOT NULL UNIQUE,
		name        TEXT    NOT NULL,
		stream_slug TEXT    NOT NULL,
		active      BOOLEAN NOT NULL DEFAULT TRUE
	)`,
}

func (s *Store) Migrate(ctx context.Context) error {
	for _, stmt := range schema {
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
