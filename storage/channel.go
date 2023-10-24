package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ChannelRepository struct {
	db *pgxpool.Pool
}

func NewChannelRepository(db *pgxpool.Pool) *ChannelRepository {
	return &ChannelRepository{db}
}

func (c *ChannelRepository) GetChannels() ([]string, error) {
	rows, err := c.db.Query(context.Background(), "SELECT name FROM channels")
	if err != nil {
		return nil, fmt.Errorf("error fetching channels: %w", err)
	}
	defer rows.Close()

	channels := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error scanning channels: %w", err)
		}
		channels = append(channels, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over channels: %w", err)
	}

	return channels, nil
}

func (c *ChannelRepository) SaveChannel(name string) error {
	_, err := c.db.Exec(context.Background(), "INSERT INTO channels (name) VALUES ($1)", name)
	if err != nil {
		return fmt.Errorf("error saving channel: %w", err)
	}
	return nil
}

func (c *ChannelRepository) GetChannel(name string) (string, error) {
	var channelName string
	err := c.db.QueryRow(context.Background(), "SELECT name FROM channels WHERE name = $1", name).Scan(&channelName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("channel with name '%s' not found", name)
		}
		return "", fmt.Errorf("error fetching channel: %w", err)
	}

	return channelName, nil
}
