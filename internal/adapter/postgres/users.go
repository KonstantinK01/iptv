package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/KonstantinK01/iptv/internal/dto"
	"github.com/jackc/pgx/v5"
)

func (s *Store) UserByLogin(ctx context.Context, login string) (dto.User, error) {
	var u dto.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, login, password_hash, active FROM users WHERE login = $1`, login,
	).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.User{}, dto.ErrNotFound
	}
	if err != nil {
		return dto.User{}, fmt.Errorf("query user: %w", err)
	}
	return u, nil
}
