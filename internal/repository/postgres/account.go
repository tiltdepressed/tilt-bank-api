// Package repository
package repository

import (
	"context"
	"errors"
	"fmt"

	"bank/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, account models.Account) (string, error)
	GetAccountByID(ctx context.Context, id string) (models.Account, error)
	GetAccountsByUserID(ctx context.Context, UserID string) ([]models.Account, error)
	GetAllAccounts(ctx context.Context) ([]models.Account, error)
	DeleteAccount(ctx context.Context, id string) error

	Deposit(ctx context.Context, id string, amount decimal.Decimal) (decimal.Decimal, error)
	Withdraw(ctx context.Context, id string, amount decimal.Decimal) (decimal.Decimal, error)
	Transfer(ctx context.Context, fromID, toID string, acmount decimal.Decimal) error
}

type accountRepo struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) CreateAccount(ctx context.Context, account models.Account) (string, error) {
	q := `
    INSERT
    INTO accounts (UserID, balance, currency) 
    VALUES ($1, $2, $3) 
    RETURNING id`

	var newID string
	err := r.db.QueryRow(ctx, q, account.UserID, account.Balance, account.Currency).Scan(&newID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return "", fmt.Errorf("user not found")
		}
		return "", err
	}

	return newID, nil
}

func (r *accountRepo) GetAccountByID(ctx context.Context, id string) (models.Account, error) {
	q := `
        SELECT id, UserID, balance, currency
        FROM accounts
        WHERE id = $1`

	var a models.Account
	err := r.db.QueryRow(ctx, q, id).Scan(&a.ID, &a.UserID, &a.Balance, &a.Currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Account{}, fmt.Errorf("account not found")
		}
		return models.Account{}, err
	}
	return a, nil
}

func (r *accountRepo) GetAccountsByUserID(ctx context.Context, UserID string) ([]models.Account, error) {
	q := `
    SELECT id, UserID, balance, currency
    FROM accounts
    WHERE UserID = $1`

	rows, err := r.db.Query(ctx, q, UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]models.Account, 0)

	for rows.Next() {
		var a models.Account

		if err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.Currency); err != nil {
			return nil, err
		}

		accounts = append(accounts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepo) GetAllAccounts(ctx context.Context) ([]models.Account, error) {
	q := `
    SELECT id, UserID, balance, currency
    FROM accounts`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]models.Account, 0)

	for rows.Next() {
		var a models.Account

		if err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.Currency); err != nil {
			return nil, err
		}

		accounts = append(accounts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepo) DeleteAccount(ctx context.Context, id string) error {
	q := `
    DELETE
    FROM accounts
    WHERE id = $1`

	cmdTag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}

func (r *accountRepo) Deposit(ctx context.Context, id string, amount decimal.Decimal) (decimal.Decimal, error) {
	q := `
    UPDATE accounts
    SET balance = balance + $2
    WHERE id = $1
    RETURNING balance`

	var b decimal.Decimal
	err := r.db.QueryRow(ctx, q, id, amount).Scan(&b)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Zero, fmt.Errorf("account not found")
		}
		return decimal.Zero, err
	}
	return b, nil
}

func (r *accountRepo) Withdraw(ctx context.Context, id string, amount decimal.Decimal) (decimal.Decimal, error) {
	q := `
    UPDATE accounts
    SET balance = balance - $2
    WHERE id = $1 AND balance >= $2
    RETURNING balance`

	var b decimal.Decimal
	err := r.db.QueryRow(ctx, q, id, amount).Scan(&b)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Zero, fmt.Errorf("account not found")
		}
		return decimal.Zero, err
	}

	return b, nil
}

func (r *accountRepo) Transfer(ctx context.Context, fromID, toID string, amount decimal.Decimal) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qWithdraw := `
    UPDATE accounts
    SET balance = balance - $2
    WHERE id = $1 AND balance >= $2`

	cmdTag, err := tx.Exec(ctx, qWithdraw, fromID, amount)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient funds or sender account not found")
	}

	qDeposit := `
    UPDATE accounts
    SET balance = balance + $2
    WHERE id = $1`

	cmdTag, err = tx.Exec(ctx, qDeposit, toID, amount)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient funds or sender account not found")
	}

	return tx.Commit(ctx)
}
