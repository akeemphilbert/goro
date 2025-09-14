package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/akeemphilbert/goro/internal/user/application"
	"github.com/akeemphilbert/goro/internal/user/domain"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// AccountHandler handles HTTP requests for account management
type AccountHandler struct {
	accountService application.AccountService
	userService    application.UserService
	logger         log.Logger
}

// NewAccountHandler creates a new AccountHandler
func NewAccountHandler(accountService application.AccountService, userService application.UserService, logger log.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		userService:    userService,
		logger:         logger,
	}
}

// CreateAccountRequest represents the HTTP request for account creation
type CreateAccountRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InviteUserRequest represents the HTTP request for user invitation
type InviteUserRequest struct {
	Email  string `json:"email"`
	RoleID string `json:"role_id"`
}

// AcceptInvitationRequest represents the HTTP request for accepting invitations
type AcceptInvitationRequest struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// UpdateMemberRoleRequest represents the HTTP request for updating member roles
type UpdateMemberRoleRequest struct {
	RoleID string `json:"role_id"`
}

// AccountResponse represents the HTTP response for account data
type AccountResponse struct {
	ID          string                 `json:"id"`
	OwnerID     string                 `json:"owner_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Settings    domain.AccountSettings `json:"settings"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// InvitationResponse represents the HTTP response for invitation data
type InvitationResponse struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Email     string `json:"email"`
	RoleID    string `json:"role_id"`
	Status    string `json:"status"`
	InvitedBy string `json:"invited_by"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

// CreateAccount handles account creation requests
func (h *AccountHandler) CreateAccount(ctx khttp.Context) error {
	// Extract owner ID from path or authentication context
	vars := ctx.Vars()
	ownerIDSlice, exists := vars["owner_id"]
	if !exists || len(ownerIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_OWNER_ID", "Owner ID is required")
	}

	ownerID := ownerIDSlice[0]
	if strings.TrimSpace(ownerID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_OWNER_ID", "Owner ID is required")
	}

	// Parse request body
	var req CreateAccountRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate input
	if err := h.validateCreateAccountRequest(req, ownerID); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	// Call service
	account, err := h.accountService.CreateAccount(ctx.Request().Context(), ownerID, req.Name)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	// Build response
	response := h.buildAccountResponse(account)

	return ctx.JSON(http.StatusCreated, response)
}

// InviteUser handles user invitation requests
func (h *AccountHandler) InviteUser(ctx khttp.Context) error {
	// Extract account ID from path
	vars := ctx.Vars()
	accountIDSlice, exists := vars["account_id"]
	if !exists || len(accountIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_ACCOUNT_ID", "Account ID is required")
	}

	accountID := accountIDSlice[0]
	if strings.TrimSpace(accountID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Account ID is required")
	}

	// Extract inviter ID from authentication context (for now, from path)
	inviterIDSlice, exists := vars["inviter_id"]
	if !exists || len(inviterIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_INVITER_ID", "Inviter ID is required")
	}

	inviterID := inviterIDSlice[0]
	if strings.TrimSpace(inviterID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_INVITER_ID", "Inviter ID is required")
	}

	// Parse request body
	var req InviteUserRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate input
	if err := h.validateInviteUserRequest(req, accountID, inviterID); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	// TODO: Check permissions - inviter must have permission to invite users
	// This would typically involve checking the inviter's role in the account

	// Call service
	invitation, err := h.accountService.InviteUser(ctx.Request().Context(), accountID, inviterID, req.Email, req.RoleID)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	// Build response
	response := h.buildInvitationResponse(invitation)

	return ctx.JSON(http.StatusCreated, response)
}

// AcceptInvitation handles invitation acceptance requests
func (h *AccountHandler) AcceptInvitation(ctx khttp.Context) error {
	// Parse request body
	var req AcceptInvitationRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate input
	if strings.TrimSpace(req.Token) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_TOKEN", "Invitation token is required")
	}

	if strings.TrimSpace(req.UserID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required")
	}

	// Call service
	err := h.accountService.AcceptInvitation(ctx.Request().Context(), req.Token, req.UserID)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Invitation accepted successfully"})
}

// UpdateMemberRole handles member role update requests
func (h *AccountHandler) UpdateMemberRole(ctx khttp.Context) error {
	// Extract account ID from path
	vars := ctx.Vars()
	accountIDSlice, exists := vars["account_id"]
	if !exists || len(accountIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_ACCOUNT_ID", "Account ID is required")
	}

	accountID := accountIDSlice[0]
	if strings.TrimSpace(accountID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Account ID is required")
	}

	// Extract user ID from path
	userIDSlice, exists := vars["user_id"]
	if !exists || len(userIDSlice) == 0 {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_USER_ID", "User ID is required")
	}

	userID := userIDSlice[0]
	if strings.TrimSpace(userID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_USER_ID", "User ID is required")
	}

	// Parse request body
	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(ctx.Request().Body).Decode(&req); err != nil {
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
	}

	// Validate input
	if strings.TrimSpace(req.RoleID) == "" {
		return h.handleError(ctx, http.StatusBadRequest, "MISSING_ROLE_ID", "Role ID is required")
	}

	// TODO: Check permissions - requester must have permission to update member roles
	// This would typically involve checking the requester's role in the account

	// Call service
	err := h.accountService.UpdateMemberRole(ctx.Request().Context(), accountID, userID, req.RoleID)
	if err != nil {
		return h.handleServiceError(ctx, err)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Member role updated successfully"})
}

// Helper methods

func (h *AccountHandler) validateCreateAccountRequest(req CreateAccountRequest, ownerID string) error {
	// Validate owner ID
	if strings.TrimSpace(ownerID) == "" {
		return fmt.Errorf("owner ID is required")
	}

	// Validate account name
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}

	return nil
}

func (h *AccountHandler) validateInviteUserRequest(req InviteUserRequest, accountID, inviterID string) error {
	// Validate account ID
	if strings.TrimSpace(accountID) == "" {
		return fmt.Errorf("account ID is required")
	}

	// Validate inviter ID
	if strings.TrimSpace(inviterID) == "" {
		return fmt.Errorf("inviter ID is required")
	}

	// Validate email
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}

	if !isValidEmailForAccount(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate role ID
	if strings.TrimSpace(req.RoleID) == "" {
		return fmt.Errorf("role ID is required")
	}

	return nil
}

func (h *AccountHandler) buildAccountResponse(account *domain.Account) AccountResponse {
	return AccountResponse{
		ID:          account.ID(),
		OwnerID:     account.OwnerID,
		Name:        account.Name,
		Description: account.Description,
		Settings:    account.Settings,
		CreatedAt:   account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   account.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *AccountHandler) buildInvitationResponse(invitation *domain.Invitation) InvitationResponse {
	return InvitationResponse{
		ID:        invitation.ID(),
		AccountID: invitation.AccountID,
		Email:     invitation.Email,
		RoleID:    invitation.RoleID,
		Status:    string(invitation.Status),
		InvitedBy: invitation.InvitedBy,
		ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt: invitation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *AccountHandler) handleError(ctx khttp.Context, status int, code, message string) error {
	response := map[string]interface{}{
		"error":   code,
		"message": message,
	}
	return ctx.JSON(status, response)
}

func (h *AccountHandler) handleServiceError(ctx khttp.Context, err error) error {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "not found"):
		return h.handleError(ctx, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case strings.Contains(errMsg, "insufficient permissions"):
		return h.handleError(ctx, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
	case strings.Contains(errMsg, "expired") || strings.Contains(errMsg, "invalid"):
		return h.handleError(ctx, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case strings.Contains(errMsg, "already"):
		return h.handleError(ctx, http.StatusConflict, "ALREADY_EXISTS", err.Error())
	default:
		h.logger.Log(log.LevelError, "msg", "Service error", "error", err.Error())
		return h.handleError(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

// isValidEmailForAccount validates email format (simple validation)
func isValidEmailForAccount(email string) bool {
	// Simple email validation - contains @ and at least one dot
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
