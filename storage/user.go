package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"investorchat/user"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db}
}

func (r *UserRepository) SaveUser(username, password string) error {
	_, err := r.db.Exec(context.Background(), "INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
	if err != nil {
		return fmt.Errorf("error saving user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUser(username string) (*user.UserModel, error) {
	var storedPassword string
	err := r.db.QueryRow(context.Background(), "SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	return &user.UserModel{
		Username: username,
		Password: storedPassword,
	}, nil
}
