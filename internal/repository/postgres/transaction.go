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

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction models.Transaction) (string, error)
	GetTransactionByID(ctx context.Context, id string) (models.Transaction, error)
	GetTransactionsByAccountID(ctx context.Context, accountID string, limit, offset int) ([]models.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id string, status models.TransactionStatus) error
}

type transactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) CreateTransaction(ctx context.Context, transaction models.Transaction) (string, error) {
	q := `
    INSERT
    INTO transactions
    (from_account_id, to_account_id, amount, status, created_at)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id`

	var newID string
	err := r.db.QueryRow(ctx, q, transaction.FromAccountID, transaction.ToAccountID, transaction.Amount, transaction.Status, transaction.CreatedAt).Scan(&newID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return "", fmt.Errorf("account not found")
		}
		return "", err
	}
	return newID, nil
}

func (r *transactionRepo) GetTransactionByID(ctx context.Context, id string) (models.Transaction, error) {
	q := `
    SELECT id, from_account_id, to_account_id, amount, status, created_at
    FROM transactions
    WHERE id = $1`

	var t models.Transaction
	err := r.db.QueryRow(ctx, q, id).Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Status, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Transaction{}, fmt.Errorf("transaction not fount")
		}
		return models.Transaction{}, err
	}
	return t, nil
}

func (r *transactionRepo) GetTransactionsByAccountID(ctx context.Context, accountID string, limit, offset int) ([]models.Transaction, error) {
	q := `
    SELECT id, from_account_id, to_account_id, amount, status, created_at
    FROM transactions
    WHERE from_account_id = $1 OR to_account_id = $1
    ORDER BY created_at DESC
    LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, q, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	txs := make([]models.Transaction, 0)

	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return txs, nil
}

func (r *transactionRepo) UpdateTransactionStatus(ctx context.Context, id string, status models.TransactionStatus) error {
	q := `
    UPDATE
    SET status = $2
    WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, q, id, status)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}
