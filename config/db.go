package config

import (
	"fmt"
	"os"
	"re-kasirpinter-go/graph/model"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDb() (*gorm.DB, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get database credentials from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "kasirpinter"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", host, user, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	err = db.AutoMigrate(
		&model.UserDB{},
		&model.UserRoleDB{},
		&model.UserPermissionDB{},
		&model.UserRolePermissionDB{},
		&model.OtpDB{},
		&model.LogEmailDB{},
		&model.LoginAuditDB{},
		&model.BlacklistedTokenDB{},
	)

	if err != nil {
		return nil, err
	}

	// Seed data
	seedUserPermissions(db)
	seedUserRoles(db)
	seedUserRolePermissions(db)
	seedAdminUser(db)

	return db, nil
}

func seedUserPermissions(db *gorm.DB) {
	permissions := []model.UserPermissionDB{
		{Name: "user.view"},
		{Name: "user.read"},
		{Name: "user.create"},
		{Name: "user.update"},
		{Name: "user.delete"},
	}

	for _, permission := range permissions {
		var existing model.UserPermissionDB
		result := db.Where("name = ?", permission.Name).First(&existing)
		if result.Error != nil {
			db.Create(&permission)
		}
	}
}

func seedUserRoles(db *gorm.DB) {
	roles := []model.UserRoleDB{
		{Name: "superadmin", IsActive: true},
		{Name: "user", IsActive: true},
	}

	for _, role := range roles {
		var existing model.UserRoleDB
		result := db.Where("name = ?", role.Name).First(&existing)
		if result.Error != nil {
			db.Create(&role)
		}
	}
}

func seedUserRolePermissions(db *gorm.DB) {
	// Get superadmin role
	var superadminRole model.UserRoleDB
	db.Where("name = ?", "superadmin").First(&superadminRole)

	// Get all permissions
	var permissions []model.UserPermissionDB
	db.Find(&permissions)

	// Assign all permissions to superadmin
	for _, permission := range permissions {
		var existing model.UserRolePermissionDB
		result := db.Where("role_id = ? AND permission_id = ?", superadminRole.ID, permission.ID).First(&existing)
		if result.Error != nil {
			rolePermission := model.UserRolePermissionDB{
				RoleID:       superadminRole.ID,
				PermissionID: permission.ID,
			}
			db.Create(&rolePermission)
		}
	}
}

func seedAdminUser(db *gorm.DB) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminRole := os.Getenv("ADMIN_ROLE")

	// Skip if admin config not set
	if adminEmail == "" || adminPassword == "" {
		fmt.Println("ADMIN_EMAIL or ADMIN_PASSWORD not set, skipping admin seed")
		return
	}

	var existingUser model.UserDB
	result := db.Where("email = ?", adminEmail).First(&existingUser)

	if result.Error == nil {
		// Admin user already exists, skip
		return
	}

	// Get superadmin role
	var adminRoleModel model.UserRoleDB
	db.Where("name = ?", adminRole).First(&adminRoleModel)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Failed to hash admin password:", err)
		return
	}

	adminUser := model.UserDB{
		Name:     "Admin",
		Email:    adminEmail,
		Password: string(hashedPassword),
		RoleID:   &adminRoleModel.ID,
		IsActive: true,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		fmt.Println("Failed to create admin user:", err)
		return
	}

	fmt.Println("Default admin user created successfully")
}
