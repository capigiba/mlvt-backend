package service

import (
	"errors"
	"mlvt/internal/entity"
	"mlvt/internal/infra/aws"
	"mlvt/internal/repo"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	RegisterUser(user *entity.User) error
	Login(email, password string) (string, uint64, error)
	ChangePassword(userID uint64, oldPassword, newPassword string) error
	UpdateUser(user *entity.User) error
	UpdateAvatar(userID uint64, avatarPath, avatarFolder string) error
	GetUserByID(userID uint64) (*entity.User, error)
	GetAllUsers() ([]entity.User, error)
	DeleteUser(userID uint64) error
	GeneratePresignedAvatarUploadURL(folder, fileName, fileType string) (string, error)
	GeneratePresignedAvatarDownloadURL(userID uint64) (string, error)
}

type userService struct {
	repo     repo.UserRepository
	s3Client aws.S3ClientInterface
	auth     AuthServiceInterface
}

func NewUserService(repo repo.UserRepository, s3Client aws.S3ClientInterface, auth AuthServiceInterface) UserService {
	return &userService{
		repo:     repo,
		s3Client: s3Client,
		auth:     auth,
	}
}

// RegisterUser creates a new user with hashed password
func (s *userService) RegisterUser(user *entity.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	user.Status = entity.UserStatusAvailable
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return s.repo.CreateUser(user)
}

// Login handles user login
func (s *userService) Login(email, password string) (string, uint64, error) {
	return s.auth.Login(email, password)
}

// ChangePassword changes a user's password
func (s *userService) ChangePassword(userID uint64, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Compare old password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return errors.New("old password does not match")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdateUserPassword(userID, string(hashedPassword))
}

// UpdateUser updates user information (except avatar)
func (s *userService) UpdateUser(user *entity.User) error {
	user.UpdatedAt = time.Now()
	return s.repo.UpdateUser(user)
}

// UpdateAvatar updates the user's avatar
func (s *userService) UpdateAvatar(userID uint64, avatarPath, avatarFolder string) error {
	return s.repo.UpdateUserAvatar(userID, avatarPath, avatarFolder)
}

// GetUserByID retrieves a user by their ID
func (s *userService) GetUserByID(userID uint64) (*entity.User, error) {
	return s.repo.GetUserByID(userID)
}

// GetAllUsers retrieves all users
func (s *userService) GetAllUsers() ([]entity.User, error) {
	return s.repo.GetAllUsers()
}

// DeleteUser soft deletes a user by setting their status to "deleted"
func (s *userService) DeleteUser(userID uint64) error {
	return s.repo.DeleteUser(userID)
}

// GeneratePresignedAvatarUploadURL generates a presigned URL for uploading an avatar
func (s *userService) GeneratePresignedAvatarUploadURL(folder, fileName, fileType string) (string, error) {
	return s.s3Client.GeneratePresignedURL(folder, fileName, fileType)
}

// GeneratePresignedAvatarDownloadURL generates a presigned URL for downloading the user's avatar
func (s *userService) GeneratePresignedAvatarDownloadURL(userID uint64) (string, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	if user.Avatar == "" || user.AvatarFolder == "" {
		return "", errors.New("avatar not found for this user")
	}

	// Generate the presigned URL for the avatar image
	url, err := s.s3Client.GeneratePresignedURL(user.AvatarFolder, user.Avatar, "image/jpeg")
	if err != nil {
		return "", err
	}

	return url, nil
}
