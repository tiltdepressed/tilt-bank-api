package repository

import (
	"context"
	"errors"
	"fmt"

	"bank/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (string, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, id string) error
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) CreateUser(ctx context.Context, user models.User) (string, error) {
	q := `
    INSERT
    INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`

	var newID string
	err := r.db.QueryRow(ctx, q, user.Username, user.PasswordHash).Scan(&newID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", fmt.Errorf("username already taken")
		}
		return "", err
	}

	return newID, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id string) (models.User, error) {
	q := `
        SELECT id, username, password_hash
        FROM users
        WHERE id = $1`

	var u models.User
	err := r.db.QueryRow(ctx, q, id).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("user not found")
		}
		return models.User{}, err
	}
	return u, nil
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	q := `
        SELECT id, username, password_hash
        FROM users
        WHERE username = $1`

	var u models.User
	err := r.db.QueryRow(ctx, q, username).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("user not found")
		}
		return models.User{}, err
	}
	return u, nil
}

func (r *userRepo) GetAllUsers(ctx context.Context) ([]models.User, error) {
	q := `
    SELECT id, username, password_hash
    FROM users`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var u models.User

		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, user models.User) error {
	q := `
    UPDATE users
    SET username = $2, password_hash = $3
    WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, q, user.ID, user.Username, user.PasswordHash)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *userRepo) DeleteUser(ctx context.Context, id string) error {
	q := `
    DELETE
    FROM users
    WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
