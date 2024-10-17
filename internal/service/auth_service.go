package service

import (
	"errors"
	"mlvt/internal/entity"
	"mlvt/internal/infra/reason"
	"mlvt/internal/infra/zap-logging/log"
	"mlvt/internal/repo"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// AuthServiceInterface defines the methods used by UserService for authentication
type AuthServiceInterface interface {
	Login(email, password string) (string, uint64, error)
	GenerateToken(user *entity.User) (string, error)
	GetUserByToken(tokenStr string) (*entity.User, error)
}

// AuthService handles user authentication
type AuthService struct {
	userRepo  repo.UserRepository
	secretKey string
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo repo.UserRepository, secretKey string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		secretKey: secretKey,
	}
}

// Login authenticates the user and returns a JWT token
func (s *AuthService) Login(email, password string) (string, uint64, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		log.Errorf("Error retrieving user by email %s: %v", email, err)
		return "", 0, errors.New(reason.UserNotFound.Message())
	}

	if user == nil {
		log.Warnf("User not found with email %s", email)
		return "", 0, errors.New(reason.UserNotFound.Message())
	}

	// Compare the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", 0, errors.New(reason.InvalidCredentials.Message())
	}

	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return "", 0, errors.New(reason.FailedToGenerateToken.Message())
	}

	return token, user.ID, nil
}

// GenerateToken creates a JWT token for a user
func (s *AuthService) GenerateToken(user *entity.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"email":  user.Email,
		"exp":    time.Now().Add(time.Hour * 72).Unix(), // Token valid for 72 hours
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserByToken extracts user information from a JWT token
func (s *AuthService) GetUserByToken(tokenStr string) (*entity.User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(reason.UnexpectedSigningMethod.Message())
		}
		return []byte(s.secretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New(reason.InvalidToken.Message())
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New(reason.InvalidTokenClaims.Message())
	}

	// Safely assert types from claims
	userIDFloat, ok := claims["userID"].(float64)
	if !ok {
		return nil, errors.New(reason.InvalidUserIDTypeInToken.Message())
	}
	userID := uint64(userIDFloat)

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New(reason.UserNotFound.Message())
	}

	return user, nil
}
