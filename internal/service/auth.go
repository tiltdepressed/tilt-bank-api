// Package service
package service

import (
	"context"
	"fmt"
	"time"

	"bank/internal/models"
	repository "bank/internal/repository/postgres"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (string, error)
	Login(ctx context.Context, username, password string) (string, error) // token + error
}

type authService struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo repository.UserRepository, jwtSecret []byte) AuthService {
	return &authService{repo: repo, jwtSecret: jwtSecret}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (s *authService) Register(ctx context.Context, username, password string) (string, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return "", err
	}
	u := models.User{
		Username:     username,
		PasswordHash: hash,
	}

	id, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (string, error) {
	u, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		// "role":     u.Role,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
