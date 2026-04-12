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
	Permissions []UserPermissionDB `gorm:"many2many:user_role_permissions;" json:"permissions,omitempty"`
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
