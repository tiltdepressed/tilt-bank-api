package service

import (
	"context"
	"fmt"

	"bank/internal/models"
	repository "bank/internal/repository/postgres"
)

type AccountService interface {
	CreateAccount(ctx context.Context, userID, currency string) (string, error) // id, err
	GetAccount(ctx context.Context, accountID, requesterID string) (models.Account, error)
}

type accountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo: repo}
}

func (s *accountService) CreateAccount(ctx context.Context, userID, currency string) (string, error) {
	a := models.Account{
		UserID:   userID,
		Currency: currency,
	}
	id, err := s.repo.CreateAccount(ctx, a)
	if err != nil {
		return "", err
	}
	return id, err
}

func (s *accountService) GetAccount(ctx context.Context, accountID, requesterID string) (models.Account, error) {
	a, err := s.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return models.Account{}, err
	}
	if a.UserID != requesterID {
		// TODO: if admin ...

		return models.Account{}, fmt.Errorf("access denied")
	}
	return a, nil
}
