package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/mosgor/Evently/backend/internal/entity"
)

type UserRepository struct {
	db *Postgres
}

func (ur *UserRepository) GetUser(ctx context.Context, id int) (*entity.User, error) {
	q := `SELECT id, name, email, password FROM users WHERE id = $1`
	var user entity.User
	err := ur.db.Pool.QueryRow(ctx, q, id).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error in getting user: %w", err)
	}
	return &user, nil
}

func (ur *UserRepository) GetByMail(ctx context.Context, email, password string) (*entity.User, error) {
	q := `SELECT id, name FROM users WHERE email = $1 AND password = $2`
	var user entity.User
	user.Password = password
	user.Email = email
	err := ur.db.Pool.QueryRow(ctx, q, email, password).Scan(
		&user.Id, &user.Name,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error in getting user: %w", err)
	}
	return &user, nil
}

func (ur *UserRepository) AddUser(ctx context.Context, user *entity.User) error {
	q := `INSERT INTO users(name, email, password) VALUES ($1, $2, $3) RETURNING id;`
	err := ur.db.Pool.QueryRow(ctx, q, user.Name, user.Email, user.Password).Scan(&user.Id)
	if err != nil {
		return fmt.Errorf("error in adding user: %w", err)
	}
	return nil
}

func (ur *UserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	q := `UPDATE users SET name = $1, email = $2, password = $3 WHERE id = $4`
	_, err := ur.db.Pool.Exec(ctx, q, user.Name, user.Email, user.Password, user.Id)
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrorNotFound
	}
	if err != nil {
		return fmt.Errorf("error in updating user: %w", err)
	}
	return nil
}

func NewUserRepository(db *Postgres) *UserRepository {
	return &UserRepository{db: db}
}
