package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/akeemphilbert/goro/internal/user/application"
	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// UserHandler handles HTTP requests for user management
type UserHandler struct {
	userService application.UserService
	logger      log.Logger
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService application.UserService, logger log.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// RegisterUserRequest represents the HTTP request for user registration
type RegisterUserRequest struct {
	Email   string             `json:"email"`
	Profile domain.UserProfile `json:"profile"`
}

// UpdateProfileRequest represents the HTTP request for profile updates
type UpdateProfileRequest struct {
	Profile domain.UserProfile `json:"profile"`
}

// DeleteAccountRequest represents the HTTP request for account deletion
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation"`
}

// UserResponse represents the HTTP response for user data
type UserResponse struct {
	ID        string             `json:"id"`
	WebID     string             `json:"webid"`
	Email     string             `json:"email"`
	Profile   domain.UserProfile `json:"profile"`
	Status    string             `json:"status"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

// RegisterUser handles user registration requests
func (h *UserHandler) RegisterUser(ctx khttp.Context) error {
	// Parse request body
	var req RegisterUserRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate input
	if err := h.validateRegistrationRequest(req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	// Call service
	appReq := application.RegisterUserRequest{
		Email:   req.Email,
		Profile: req.Profile,
	}

	user, err := h.userService.RegisterUser(ctx.Request().Context(), appReq)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	// Build response
	response := h.buildUserResponse(user)

	return ctx.JSON(http.StatusCreated, response)
}

// GetUser handles get user requests
func (h *UserHandler) GetUser(ctx khttp.Context) error {
	// Extract user ID from path
	vars := ctx.Vars()
	userIDSlice, exists := vars["id"]
	if !exists || len(userIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required")
	}

	userID := userIDSlice[0]
	if strings.TrimSpace(userID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_USER_ID", "User ID is required")
	}

	// Call service
	user, err := h.userService.GetUserByID(ctx.Request().Context(), userID)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	// Build response
	response := h.buildUserResponse(user)

	return ctx.JSON(http.StatusOK, response)
}

// UpdateProfile handles profile update requests
func (h *UserHandler) UpdateProfile(ctx khttp.Context) error {
	// Extract user ID from path
	vars := ctx.Vars()
	userIDSlice, exists := vars["id"]
	if !exists || len(userIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required")
	}

	userID := userIDSlice[0]
	if strings.TrimSpace(userID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_USER_ID", "User ID is required")
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate profile
	if err := req.Profile.Validate(); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	// Call service
	err := h.userService.UpdateProfile(ctx.Request().Context(), userID, req.Profile)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Profile updated successfully"})
}

// DeleteAccount handles account deletion requests
func (h *UserHandler) DeleteAccount(ctx khttp.Context) error {
	// Extract user ID from path
	vars := ctx.Vars()
	userIDSlice, exists := vars["id"]
	if !exists || len(userIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required")
	}

	userID := userIDSlice[0]
	if strings.TrimSpace(userID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_USER_ID", "User ID is required")
	}

	// Parse request body
	var req DeleteAccountRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate confirmation
	if strings.TrimSpace(req.Confirmation) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_CONFIRMATION", "Confirmation is required")
	}

	if req.Confirmation != "DELETE" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_CONFIRMATION", "Invalid confirmation. Must be 'DELETE'")
	}

	// Call service
	err := h.userService.DeleteAccount(ctx.Request().Context(), userID)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	return ctx.JSON(http.StatusNoContent, nil)
}

// Helper methods

func (h *UserHandler) validateRegistrationRequest(req RegisterUserRequest) error {
	// Validate email
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if !isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate profile
	if err := req.Profile.Validate(); err != nil {
		return err
	}

	return nil
}

func (h *UserHandler) buildUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID(),
		WebID:     user.WebID,
		Email:     user.Email,
		Profile:   user.Profile,
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *UserHandler) handleError(ctx khttp.Context, status int, code, message string) error {
	response := map[string]interface{}{
		"error":   code,
		"message": message,
	}
	return ctx.JSON(status, response)
}

func (h *UserHandler) handleServiceError(ctx khttp.Context, err error) error {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "not found"):
		return h.handleError(ctx, http.StatusNotFound, "NOT_FOUND", "User not found")
	case strings.Contains(errMsg, "already exists"):
		return h.handleError(ctx, http.StatusConflict, "ALREADY_EXISTS", "User already exists")
	case strings.Contains(errMsg, "invalid"):
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	default:
		h.logger.Log(log.LevelError, "msg", "Service error", "error", err.Error())
		return h.handleError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
