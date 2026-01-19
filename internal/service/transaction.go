package service

import (
	"context"
	"fmt"

	"bank/internal/models"
	repository "bank/internal/repository/postgres"

	"github.com/shopspring/decimal"
)

type TransactionService interface {
	Transfer(ctx context.Context, userID, fromAccountID, toAccountID string, amount decimal.Decimal) (string, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
) TransactionService {
	return &transactionService{transactionRepo: transactionRepo, accountRepo: accountRepo}
}

func (s *transactionService) Transfer(ctx context.Context, userID, fromAccountID, toAccountID string, amount decimal.Decimal) (string, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return "", fmt.Errorf("invalid amount")
	}
	if fromAccountID == toAccountID {
		return "", fmt.Errorf("cannot transfer to self")
	}

	fromAccount, err := s.accountRepo.GetAccountByID(ctx, fromAccountID)
	if err != nil {
		return "", err
	}

	if fromAccount.UserID != userID {
		return "", fmt.Errorf("access denied to source account")
	}

	toAccount, err := s.accountRepo.GetAccountByID(ctx, toAccountID)
	if err != nil {
		return "", err
	}

	if fromAccount.Currency != toAccount.Currency {
		return "", fmt.Errorf("currenciy mismatch: %s vs %s", fromAccount.Currency, toAccount.Currency)
	}

	err = s.accountRepo.Transfer(ctx, fromAccountID, toAccountID, amount)
	if err != nil {
		return "", err
	}

	t := models.Transaction{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Status:        models.StatusSuccess,
	}

	transactionID, err := s.transactionRepo.CreateTransaction(ctx, t)
	if err != nil {
		return "", fmt.Errorf("transfer successful, but failed to save history: %w", err)
	}

	return transactionID, nil
}
