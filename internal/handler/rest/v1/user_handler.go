package handler

import (
	"net/http"
	"strconv"

	"mlvt/internal/entity"
	"mlvt/internal/infra/env"
	"mlvt/internal/pkg/response"
	"mlvt/internal/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

// RegisterUser godoc
// @Summary Register a new user
// @Description Creates a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body entity.User true "User data"
// @Success 201 {object} response.MessageResponse "message"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/register [post]
func (h *UserController) RegisterUser(c *gin.Context) {
	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.userService.RegisterUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response.MessageResponse{Message: "User registered successfully"})
}

// GenerateAvatarDownloadURL godoc
// @Summary Get presigned URL for avatar download
// @Description Generates a presigned URL to download the user's avatar from S3
// @Tags users
// @Produce json
// @Param user_id path uint64 true "User ID"
// @Success 200 {object} response.AvatarDownloadURLResponse "avatar_download_url"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 404 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/{user_id}/avatar-download-url [get]
func (h *UserController) GenerateAvatarDownloadURL(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid user ID"})
		return
	}

	// Call the service to generate the presigned download URL for the avatar
	url, err := h.userService.GeneratePresignedAvatarDownloadURL(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.AvatarDownloadURLResponse{AvatarDownloadURL: url})
}

// LoginUser godoc
// @Summary User login
// @Description Logs in a user with their email and password
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body object true "Email and password"
// @Success 200 {object} response.TokenResponse "token"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 401 {object} response.ErrorResponse "error"
// @Router /users/login [post]
func (h *UserController) LoginUser(c *gin.Context) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	token, userID, err := h.userService.Login(credentials.Email, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.TokenResponse{
		Token:  token,
		UserID: userID,
	})
}

// ChangePassword godoc
// @Summary Change user password
// @Description Allows a user to change their password
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path uint64 true "User ID"
// @Param password body object true "Old and new password"
// @Success 200 {object} response.MessageResponse "message"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/{user_id}/change-password [put]
func (h *UserController) ChangePassword(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var request struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid input"})
		return
	}

	if err := h.userService.ChangePassword(userID, request.OldPassword, request.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "Password changed successfully"})
}

// UpdateUser godoc
// @Summary Update user information
// @Description Updates the user's information, excluding the avatar
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path uint64 true "User ID"
// @Param user body entity.User true "User data"
// @Success 200 {object} response.MessageResponse "message"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/{user_id} [put]
func (h *UserController) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}
	user.ID = userID

	if err := h.userService.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "User updated successfully"})
}

// UpdateAvatar godoc
// @Summary Update user avatar
// @Description Generates a presigned URL for uploading the user's avatar
// @Tags users
// @Produce json
// @Param user_id path uint64 true "User ID"
// @Param file_name query string true "File name for avatar"
// @Success 200 {object} response.AvatarUploadURLResponse "avatar_upload_url"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/{user_id}/update-avatar [put]
func (h *UserController) UpdateAvatar(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid user ID"})
		return
	}
	fileName := c.Query("file_name")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "file_name is required"})
		return
	}

	url, err := h.userService.GeneratePresignedAvatarUploadURL(env.EnvConfig.AvatarFolder, fileName, "image/jpeg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Update the avatar path and folder in the database after a successful upload
	if err := h.userService.UpdateAvatar(userID, fileName, env.EnvConfig.AvatarFolder); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.AvatarUploadURLResponse{AvatarUploadURL: url})
}

// LoadAvatar godoc
// @Summary Load user avatar
// @Description Redirects the client to the presigned URL to download the user's avatar
// @Tags users
// @Produce json
// @Param user_id path uint64 true "User ID"
// @Success 307 {string} string "Redirects to avatar URL"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/{user_id}/avatar [get]
func (h *UserController) LoadAvatar(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid user ID"})
		return
	}

	// Call the service to generate the presigned download URL for the avatar
	url, err := h.userService.GeneratePresignedAvatarDownloadURL(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Redirect the user to the presigned URL
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Fetches a user's details by their ID
// @Tags users
// @Produce json
// @Param user_id path uint64 true "User ID"
// @Success 200 {object} response.UserResponse "user"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 404 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/{user_id} [get]
func (h *UserController) GetUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.UserResponse{User: *user})
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Retrieves a list of all users in the system
// @Tags users
// @Produce json
// @Success 200 {object} response.UsersResponse "users"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users [get]
func (h *UserController) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, response.UsersResponse{Users: users})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft deletes a user by updating their status
// @Tags users
// @Produce json
// @Param user_id path uint64 true "User ID"
// @Success 200 {object} response.MessageResponse "message"
// @Failure 400 {object} response.ErrorResponse "error"
// @Failure 404 {object} response.ErrorResponse "error"
// @Failure 500 {object} response.ErrorResponse "error"
// @Router /users/{user_id} [delete]
func (h *UserController) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid user ID"})
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "User deleted successfully"})
}
