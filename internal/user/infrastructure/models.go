package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// UserModel represents the GORM model for users table
type UserModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(255)"`
	WebID     string    `gorm:"uniqueIndex;not null;type:varchar(500)"`
	Email     string    `gorm:"uniqueIndex;not null;type:varchar(255)"`
	Name      string    `gorm:"type:varchar(255)"`
	Status    string    `gorm:"not null;type:varchar(50)"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName specifies the table name for UserModel
func (UserModel) TableName() string {
	return "user_models"
}

// UpdateStatus updates the user status with logging
func (u *UserModel) UpdateStatus(ctx context.Context, status string) error {
	log.Context(ctx).Debugf("[UserModel.UpdateStatus] Updating user status: userID=%s, currentStatus=%s, newStatus=%s", u.ID, u.Status, status)

	if status == "" {
		err := fmt.Errorf("status cannot be empty")
		log.Context(ctx).Debug("[UserModel.UpdateStatus] Validation failed: status cannot be empty")
		return err
	}

	u.Status = status
	u.UpdatedAt = time.Now()
	log.Context(ctx).Debugf("[UserModel.UpdateStatus] Status updated successfully: userID=%s, status=%s", u.ID, status)
	return nil
}

// UpdateProfile updates the user profile information with logging
func (u *UserModel) UpdateProfile(ctx context.Context, name string) error {
	log.Context(ctx).Debugf("[UserModel.UpdateProfile] Updating user profile: userID=%s, currentName=%s, newName=%s", u.ID, u.Name, name)

	if len(name) > 255 {
		err := fmt.Errorf("name must be less than 256 characters")
		log.Context(ctx).Debug("[UserModel.UpdateProfile] Validation failed: name too long")
		return err
	}

	u.Name = name
	u.UpdatedAt = time.Now()
	log.Context(ctx).Debugf("[UserModel.UpdateProfile] Profile updated successfully: userID=%s, name=%s", u.ID, name)
	return nil
}

// RoleModel represents the GORM model for roles table
type RoleModel struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	Name        string    `gorm:"not null;type:varchar(255)"`
	Description string    `gorm:"type:text"`
	Permissions string    `gorm:"type:text"` // JSON serialized permissions
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// TableName specifies the table name for RoleModel
func (RoleModel) TableName() string {
	return "role_models"
}

// UpdateRole updates the role information with logging
func (r *RoleModel) UpdateRole(ctx context.Context, name, description, permissions string) error {
	log.Context(ctx).Debugf("[RoleModel.UpdateRole] Updating role: roleID=%s, currentName=%s, newName=%s", r.ID, r.Name, name)

	if name == "" {
		err := fmt.Errorf("role name cannot be empty")
		log.Context(ctx).Debug("[RoleModel.UpdateRole] Validation failed: name cannot be empty")
		return err
	}

	if len(name) > 255 {
		err := fmt.Errorf("role name must be less than 256 characters")
		log.Context(ctx).Debug("[RoleModel.UpdateRole] Validation failed: name too long")
		return err
	}

	r.Name = name
	r.Description = description
	r.Permissions = permissions
	r.UpdatedAt = time.Now()

	log.Context(ctx).Debugf("[RoleModel.UpdateRole] Role updated successfully: roleID=%s, name=%s", r.ID, name)
	return nil
}

// AccountModel represents the GORM model for accounts table
type AccountModel struct {
	ID          string    `gorm:"primaryKey;type:varchar(255)"`
	OwnerID     string    `gorm:"not null;type:varchar(255);index"`
	Owner       UserModel `gorm:"foreignKey:OwnerID;references:ID;constraint:OnDelete:CASCADE"`
	Name        string    `gorm:"not null;type:varchar(255)"`
	Description string    `gorm:"type:text"`
	Settings    string    `gorm:"type:text"` // JSON serialized AccountSettings
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// TableName specifies the table name for AccountModel
func (AccountModel) TableName() string {
	return "account_models"
}

// UpdateAccount updates the account information with logging
func (a *AccountModel) UpdateAccount(ctx context.Context, name, description, settings string) error {
	log.Context(ctx).Debugf("[AccountModel.UpdateAccount] Updating account: accountID=%s, currentName=%s, newName=%s", a.ID, a.Name, name)

	if name == "" {
		err := fmt.Errorf("account name cannot be empty")
		log.Context(ctx).Debug("[AccountModel.UpdateAccount] Validation failed: name cannot be empty")
		return err
	}

	if len(name) > 255 {
		err := fmt.Errorf("account name must be less than 256 characters")
		log.Context(ctx).Debug("[AccountModel.UpdateAccount] Validation failed: name too long")
		return err
	}

	a.Name = name
	a.Description = description
	a.Settings = settings
	a.UpdatedAt = time.Now()

	log.Context(ctx).Debugf("[AccountModel.UpdateAccount] Account updated successfully: accountID=%s, name=%s", a.ID, name)
	return nil
}

// TransferOwnership transfers account ownership with logging
func (a *AccountModel) TransferOwnership(ctx context.Context, newOwnerID string) error {
	log.Context(ctx).Debugf("[AccountModel.TransferOwnership] Transferring account ownership: accountID=%s, currentOwner=%s, newOwner=%s", a.ID, a.OwnerID, newOwnerID)

	if newOwnerID == "" {
		err := fmt.Errorf("new owner ID cannot be empty")
		log.Context(ctx).Debug("[AccountModel.TransferOwnership] Validation failed: new owner ID cannot be empty")
		return err
	}

	oldOwnerID := a.OwnerID
	a.OwnerID = newOwnerID
	a.UpdatedAt = time.Now()

	log.Context(ctx).Infof("Account ownership transferred: accountID=%s, from=%s, to=%s", a.ID, oldOwnerID, newOwnerID)
	return nil
}

// AccountMemberModel represents the GORM model for account membership (projection from events)
type AccountMemberModel struct {
	ID        string       `gorm:"primaryKey;type:varchar(255)"`
	AccountID string       `gorm:"not null;type:varchar(255);index;uniqueIndex:idx_account_user"`
	Account   AccountModel `gorm:"foreignKey:AccountID;references:ID;constraint:OnDelete:CASCADE"`
	UserID    string       `gorm:"not null;type:varchar(255);index;uniqueIndex:idx_account_user"`
	User      UserModel    `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	RoleID    string       `gorm:"not null;type:varchar(255)"`
	Role      RoleModel    `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE"`
	InvitedBy string       `gorm:"type:varchar(255)"`
	Inviter   *UserModel   `gorm:"foreignKey:InvitedBy;references:ID;constraint:OnDelete:SET NULL"`
	JoinedAt  time.Time    `gorm:"not null"`
	CreatedAt time.Time    `gorm:"not null"`
	UpdatedAt time.Time    `gorm:"not null"`
}

// TableName specifies the table name for AccountMemberModel
func (AccountMemberModel) TableName() string {
	return "account_member_models"
}

// UpdateRole updates the member's role with logging
func (m *AccountMemberModel) UpdateRole(ctx context.Context, newRoleID string) error {
	log.Context(ctx).Debugf("[AccountMemberModel.UpdateRole] Updating member role: memberID=%s, accountID=%s, userID=%s, currentRole=%s, newRole=%s",
		m.ID, m.AccountID, m.UserID, m.RoleID, newRoleID)

	if newRoleID == "" {
		err := fmt.Errorf("role ID cannot be empty")
		log.Context(ctx).Debug("[AccountMemberModel.UpdateRole] Validation failed: role ID cannot be empty")
		return err
	}

	oldRoleID := m.RoleID
	m.RoleID = newRoleID
	m.UpdatedAt = time.Now()

	log.Context(ctx).Infof("Member role updated: memberID=%s, accountID=%s, userID=%s, from=%s, to=%s",
		m.ID, m.AccountID, m.UserID, oldRoleID, newRoleID)
	return nil
}

// UpdateInviter updates the inviter information with logging
func (m *AccountMemberModel) UpdateInviter(ctx context.Context, inviterID string) error {
	log.Context(ctx).Debugf("[AccountMemberModel.UpdateInviter] Updating member inviter: memberID=%s, currentInviter=%s, newInviter=%s",
		m.ID, m.InvitedBy, inviterID)

	m.InvitedBy = inviterID
	m.UpdatedAt = time.Now()

	log.Context(ctx).Debugf("[AccountMemberModel.UpdateInviter] Member inviter updated: memberID=%s, inviterID=%s", m.ID, inviterID)
	return nil
}

// InvitationModel represents the GORM model for invitations table
type InvitationModel struct {
	ID        string       `gorm:"primaryKey;type:varchar(255)"`
	AccountID string       `gorm:"not null;type:varchar(255);index"`
	Account   AccountModel `gorm:"foreignKey:AccountID;references:ID;constraint:OnDelete:CASCADE"`
	Email     string       `gorm:"not null;type:varchar(255)"`
	RoleID    string       `gorm:"not null;type:varchar(255)"`
	Role      RoleModel    `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE"`
	Token     string       `gorm:"uniqueIndex;not null;type:varchar(255)"`
	InvitedBy string       `gorm:"not null;type:varchar(255)"`
	Inviter   UserModel    `gorm:"foreignKey:InvitedBy;references:ID;constraint:OnDelete:CASCADE"`
	Status    string       `gorm:"not null;type:varchar(50)"` // pending, accepted, expired, revoked
	ExpiresAt time.Time    `gorm:"not null"`
	CreatedAt time.Time    `gorm:"not null"`
	UpdatedAt time.Time    `gorm:"not null"`
}

// TableName specifies the table name for InvitationModel
func (InvitationModel) TableName() string {
	return "invitation_models"
}

// UpdateStatus updates the invitation status with logging
func (i *InvitationModel) UpdateStatus(ctx context.Context, status string) error {
	log.Context(ctx).Debugf("[InvitationModel.UpdateStatus] Updating invitation status: invitationID=%s, currentStatus=%s, newStatus=%s",
		i.ID, i.Status, status)

	if status == "" {
		err := fmt.Errorf("status cannot be empty")
		log.Context(ctx).Debug("[InvitationModel.UpdateStatus] Validation failed: status cannot be empty")
		return err
	}

	validStatuses := []string{"pending", "accepted", "expired", "revoked"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		err := fmt.Errorf("invalid status: %s", status)
		log.Context(ctx).Debugf("[InvitationModel.UpdateStatus] Validation failed: invalid status %s", status)
		return err
	}

	oldStatus := i.Status
	i.Status = status
	i.UpdatedAt = time.Now()

	log.Context(ctx).Infof("Invitation status updated: invitationID=%s, from=%s, to=%s", i.ID, oldStatus, status)
	return nil
}

// Accept marks the invitation as accepted with logging
func (i *InvitationModel) Accept(ctx context.Context) error {
	log.Context(ctx).Debugf("[InvitationModel.Accept] Accepting invitation: invitationID=%s, currentStatus=%s", i.ID, i.Status)

	if i.Status != "pending" {
		err := fmt.Errorf("only pending invitations can be accepted")
		log.Context(ctx).Debugf("[InvitationModel.Accept] Validation failed: invitation status is %s", i.Status)
		return err
	}

	return i.UpdateStatus(ctx, "accepted")
}

// Revoke marks the invitation as revoked with logging
func (i *InvitationModel) Revoke(ctx context.Context) error {
	log.Context(ctx).Debugf("[InvitationModel.Revoke] Revoking invitation: invitationID=%s, currentStatus=%s", i.ID, i.Status)

	if i.Status == "accepted" {
		err := fmt.Errorf("accepted invitations cannot be revoked")
		log.Context(ctx).Debug("[InvitationModel.Revoke] Validation failed: invitation already accepted")
		return err
	}

	return i.UpdateStatus(ctx, "revoked")
}

// MarkExpired marks the invitation as expired with logging
func (i *InvitationModel) MarkExpired(ctx context.Context) error {
	log.Context(ctx).Debugf("[InvitationModel.MarkExpired] Marking invitation as expired: invitationID=%s", i.ID)
	return i.UpdateStatus(ctx, "expired")
}
