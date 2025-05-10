package usecase

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/perfect1337/forum-service/internal/config"
	"github.com/perfect1337/forum-service/internal/repository"
)

type AuthUseCase struct {
	repo      repository.Postgres
	secretKey []byte
}

func NewAuthUseCase(repo repository.Postgres, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{
		repo:      repo,
		secretKey: []byte(cfg.Auth.SecretKey),
	}
}

// SecretKey returns the secret key used for signing and validating tokens.
func (uc *AuthUseCase) SecretKey() []byte {
	return uc.secretKey
}

// GenerateToken creates a JWT token with the specified claims.
func (uc *AuthUseCase) GenerateToken(userID int, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.SecretKey())
}

// ParseToken parses and validates a JWT token, extracting the user ID and username from the claims.
func (uc *AuthUseCase) ParseToken(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return uc.SecretKey(), nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, "", fmt.Errorf("invalid user_id in token")
		}

		username, ok := claims["username"].(string)
		if !ok {
			return 0, "", fmt.Errorf("invalid username in token")
		}

		return int64(userID), username, nil
	}

	return 0, "", fmt.Errorf("invalid token")
}
