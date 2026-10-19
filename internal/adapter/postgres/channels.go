package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/KonstantinK01/iptv/internal/dto"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListActiveChannels(ctx context.Context) ([]dto.Channel, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, number, name, stream_slug, active FROM channels WHERE active ORDER BY number`)
	if err != nil {
		return nil, fmt.Errorf("query channels: %w", err)
	}
	defer rows.Close()

	channels := []dto.Channel{}
	for rows.Next() {
		var c dto.Channel
		if err := rows.Scan(&c.ID, &c.Number, &c.Name, &c.StreamSlug, &c.Active); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate channels: %w", err)
	}
	return channels, nil
}

func (s *Store) ActiveChannelByID(ctx context.Context, id int64) (dto.Channel, error) {
	var c dto.Channel
	err := s.pool.QueryRow(ctx,
		`SELECT id, number, name, stream_slug, active FROM channels WHERE id = $1 AND active`, id,
	).Scan(&c.ID, &c.Number, &c.Name, &c.StreamSlug, &c.Active)
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.Channel{}, dto.ErrNotFound
	}
	if err != nil {
		return dto.Channel{}, fmt.Errorf("query channel: %w", err)
	}
	return c, nil
}
