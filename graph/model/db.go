package model

import (
	"time"

	"gorm.io/gorm"
)

// UserDB represents the database model for User
type UserDB struct {
	ID        int32      `gorm:"primaryKey;autoIncrement" json:"id"`
	SecureID  *string    `gorm:"uniqueIndex" json:"secure_id,omitempty"`
	Name      string     `gorm:"not null" json:"name"`
	Email     string     `gorm:"uniqueIndex;not null" json:"email"`
	Address   string     `gorm:"not null" json:"address"`
	Phone     string     `gorm:"not null" json:"phone"`
	Password  string     `gorm:"not null" json:"-"` // Never expose password in JSON
	Avatar    *string    `json:"avatar,omitempty"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	RoleID    *int64     `json:"role_id,omitempty"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Relations
	Role *UserRoleDB `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// UserRoleDB represents the database model for UserRole
type UserRoleDB struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"uniqueIndex;not null" json:"name"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy *string    `json:"created_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	DeletedBy *string    `json:"deleted_by,omitempty"`

	// Relations
	Permissions []UserPermissionDB `gorm:"many2many:user_role_permissions;foreignKey:ID;joinForeignKey:RoleID;joinReferences:PermissionID" json:"permissions,omitempty"`
}

// UserPermissionDB represents the database model for UserPermission
type UserPermissionDB struct {
	ID   int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"`
}

// UserRolePermissionDB represents the many-to-many relationship between UserRole and UserPermission
type UserRolePermissionDB struct {
	RoleID       int64 `gorm:"primaryKey" json:"role_id"`
	PermissionID int64 `gorm:"primaryKey" json:"permission_id"`

	// Relations
	Role       *UserRoleDB       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission *UserPermissionDB `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// TableName specifies the table name for UserDB
func (UserDB) TableName() string {
	return "users"
}

// TableName specifies the table name for UserRoleDB
func (UserRoleDB) TableName() string {
	return "user_roles"
}

// TableName specifies the table name for UserPermissionDB
func (UserPermissionDB) TableName() string {
	return "user_permissions"
}

// TableName specifies the table name for UserRolePermissionDB
func (UserRolePermissionDB) TableName() string {
	return "user_role_permissions"
}

// OtpDB represents the database model for OTP
type OtpDB struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"not null" json:"email"`
	Code      string    `gorm:"size:6;not null" json:"code"`
	IsValid   bool      `gorm:"default:true" json:"is_valid"`
	ExpiredAt time.Time `gorm:"not null" json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `gorm:"not null" json:"type"`
}

// TableName specifies the table name for OtpDB
func (OtpDB) TableName() string {
	return "otps"
}

// BeforeCreate hook for OtpDB
func (o *OtpDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	o.CreatedAt = now
	return nil
}

// BeforeCreate hook for UserDB
func (u *UserDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for UserDB
func (u *UserDB) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook for UserRoleDB
func (r *UserRoleDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for UserRoleDB
func (r *UserRoleDB) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// LogEmailDB represents the database model for LogEmail
type LogEmailDB struct {
	ID      int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email   string    `gorm:"not null" json:"email"`
	Action  string    `gorm:"not null" json:"action"`
	Status  string    `gorm:"not null" json:"status"`
	Message string    `gorm:"not null" json:"message"`
	Ts      time.Time `gorm:"not null" json:"ts"`
	IP      *string   `json:"ip,omitempty"`
	Browser *string   `json:"browser,omitempty"`
	OS      *string   `json:"os,omitempty"`
}

// TableName specifies the table name for LogEmailDB
func (LogEmailDB) TableName() string {
	return "log_emails"
}

// BeforeCreate hook for LogEmailDB
func (l *LogEmailDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	l.Ts = now
	return nil
}

// LoginAuditDB represents the database model for LoginAudit
type LoginAuditDB struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"not null" json:"email"`
	Success   bool      `gorm:"not null" json:"success"`
	IP        *string   `json:"ip,omitempty"`
	Browser   *string   `json:"browser,omitempty"`
	OS        *string   `json:"os,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for LoginAuditDB
func (LoginAuditDB) TableName() string {
	return "login_audits"
}

// BeforeCreate hook for LoginAuditDB
func (l *LoginAuditDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	l.CreatedAt = now
	return nil
}

// BlacklistedTokenDB represents the database model for blacklisted tokens
type BlacklistedTokenDB struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null;size:500" json:"token"`
	UserID    int32     `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expired_at"`
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`
}

// TableName specifies the table name for BlacklistedTokenDB
func (BlacklistedTokenDB) TableName() string {
	return "blacklisted_tokens"
}

// BeforeCreate hook for BlacklistedTokenDB
func (b *BlacklistedTokenDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	b.CreatedAt = now
	return nil
}
