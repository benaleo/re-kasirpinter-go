package service

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"re-kasirpinter-go/graph/input"
	"re-kasirpinter-go/graph/model"
	"re-kasirpinter-go/helper"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		DB: db,
	}
}

type ClientInfo struct {
	IP      string
	Browser string
	OS      string
}

type Claims struct {
	UserID   int32  `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	SecureID string `json:"secure_id"`
	Purpose  string `json:"purpose"`
	jwt.RegisteredClaims
}

// Login handles user authentication
func (s *AuthService) Login(ctx context.Context, input input.LoginInput, clientInfo *ClientInfo) (*model.AuthResponse, error) {
	var ip, browser, os *string
	if clientInfo != nil {
		ip = &clientInfo.IP
		browser = &clientInfo.Browser
		os = &clientInfo.OS
	}

	// Check for recent failed login attempts (within last 5 minutes)
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	var failedAttemptsCount int64
	s.DB.Model(&model.LoginAuditDB{}).
		Where("email = ? AND success = ? AND created_at > ?", input.Email, false, fiveMinutesAgo).
		Count(&failedAttemptsCount)

	// If 3 or more failed attempts in last 5 minutes, block login
	if failedAttemptsCount >= 3 {
		// Record this blocked attempt
		loginAudit := model.LoginAuditDB{
			Email:   input.Email,
			Success: false,
			IP:      ip,
			Browser: browser,
			OS:      os,
		}
		s.DB.Create(&loginAudit)

		// Calculate when user can try again (5 minutes from now)
		tryAgainAt := time.Now().Add(5 * time.Minute).Format("2006-01-02 15:04:05")

		return &model.AuthResponse{
			Code:    429,
			Success: false,
			Message: fmt.Sprintf("too many failed login attempts. please wait 5 minutes before trying again [[%s]]", tryAgainAt),
		}, nil
	}

	// Find user by email
	var userDB model.UserDB
	result := s.DB.Where("email = ? AND is_active = ?", input.Email, true).First(&userDB)
	if result.Error != nil {
		// Record failed attempt (user not found)
		loginAudit := model.LoginAuditDB{
			Email:   input.Email,
			Success: false,
			IP:      ip,
			Browser: browser,
			OS:      os,
		}
		s.DB.Create(&loginAudit)

		return &model.AuthResponse{
			Code:    401,
			Success: false,
			Message: "invalid email or password",
		}, nil
	}

	// Decrypt password from frontend using AES
	decryptedPassword, err := helper.Decrypt(input.Password)
	if err != nil {
		// Record failed attempt (decryption error)
		loginAudit := model.LoginAuditDB{
			Email:   input.Email,
			Success: false,
			IP:      ip,
			Browser: browser,
			OS:      os,
		}
		s.DB.Create(&loginAudit)

		return &model.AuthResponse{
			Code:    401,
			Success: false,
			Message: "invalid email or password",
		}, nil
	}

	// Check password
	if !helper.CheckPassword(decryptedPassword, userDB.Password) {
		// Record failed attempt (wrong password)
		loginAudit := model.LoginAuditDB{
			Email:   input.Email,
			Success: false,
			IP:      ip,
			Browser: browser,
			OS:      os,
		}
		s.DB.Create(&loginAudit)

		return &model.AuthResponse{
			Code:    401,
			Success: false,
			Message: "invalid email or password",
		}, nil
	}

	// Get user role
	var userRole model.UserRoleDB
	if userDB.RoleID != nil {
		s.DB.First(&userRole, *userDB.RoleID)
	}

	// Generate JWT token
	roleName := ""
	if userRole.ID > 0 {
		roleName = userRole.Name
	}

	secureID := ""
	if userDB.SecureID != nil {
		secureID = *userDB.SecureID
	}

	token, err := s.generateJWT(userDB.ID, userDB.Email, roleName, secureID, "login")
	if err != nil {
		return &model.AuthResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to generate token: %v", err),
		}, nil
	}

	// Store token in ActiveTokenDB with 1 hour expiry and device info
	expiryTime := time.Now().Add(1 * time.Hour)
	activeToken := model.ActiveTokenDB{
		Token:     token,
		UserID:    userDB.ID,
		IP:        ip,
		Browser:   browser,
		OS:        os,
		ExpiresAt: expiryTime,
	}
	if err := s.DB.Create(&activeToken).Error; err != nil {
		return &model.AuthResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to store token: %v", err),
		}, nil
	}

	// Record successful login attempt
	loginAudit := model.LoginAuditDB{
		Email:   input.Email,
		Success: true,
		IP:      ip,
		Browser: browser,
		OS:      os,
	}
	s.DB.Create(&loginAudit)

	// Convert DB model to GraphQL model using mapper
	var userRoleDB *model.UserRoleDB
	if userRole.ID > 0 {
		userRoleDB = &userRole
	}
	user := helper.ToGraphQLUser(userDB, userRoleDB)

	return &model.AuthResponse{
		Code:    200,
		Success: true,
		Message: "login successful",
		Data: &model.AuthData{
			Token: token,
			User:  user,
		},
	}, nil
}

// Logout handles user logout by blacklisting the token
func (s *AuthService) Logout(ctx context.Context, token string, userClaims *Claims) (*model.LogoutResponse, error) {
	if userClaims == nil {
		return &model.LogoutResponse{
			Code:    401,
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	if token == "" {
		return &model.LogoutResponse{
			Code:    400,
			Success: false,
			Message: "Token not found",
		}, nil
	}

	// Validate the token to get its expiration time
	claims, err := s.validateJWT(token)
	if err != nil {
		return &model.LogoutResponse{
			Code:    400,
			Success: false,
			Message: "Invalid token",
		}, nil
	}

	// Verify that the token belongs to the authenticated user
	if claims.UserID != userClaims.UserID {
		return &model.LogoutResponse{
			Code:    403,
			Success: false,
			Message: "Token does not belong to authenticated user",
		}, nil
	}

	// Get the token's expiration time
	expiresAt := claims.ExpiresAt.Time

	// Add the token to the blacklist
	err = s.blacklistToken(token, userClaims.UserID, expiresAt)
	if err != nil {
		return &model.LogoutResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("Failed to blacklist token: %v", err),
		}, nil
	}

	// Optionally cleanup expired blacklisted tokens
	go s.cleanupExpiredBlacklistedTokens()

	return &model.LogoutResponse{
		Code:    200,
		Success: true,
		Message: "logout successful",
	}, nil
}

// CreateOtp creates and sends an OTP code to the user's email
func (s *AuthService) CreateOtp(ctx context.Context, input model.CreateOtpInput, clientInfo *ClientInfo, enqueueEmailJob func(db *gorm.DB, email, code string, retry bool, ip, browser, os *string) error) (*model.CreateOtpResponse, error) {
	// Check if user exists
	var userDB model.UserDB
	result := s.DB.Where("email = ? AND is_active = ?", input.Email, true).First(&userDB)
	if result.Error != nil {
		return &model.CreateOtpResponse{
			Code:    404,
			Success: false,
			Message: "User not found",
		}, nil
	}

	// Check if there's a valid OTP created recently (within 30 seconds)
	var existingOTP model.OtpDB
	thirtySecondsAgo := time.Now().Add(-30 * time.Second)
	result = s.DB.Where("email = ? AND type = ? AND is_valid = ? AND created_at > ?", input.Email, input.Type, true, thirtySecondsAgo).Order("created_at DESC").First(&existingOTP)

	if result.Error == nil {
		// Valid OTP exists recently, return it instead of creating a new one
		var ip, browser, os *string
		if clientInfo != nil {
			ip = &clientInfo.IP
			browser = &clientInfo.Browser
			os = &clientInfo.OS
		}

		// Determine retry value
		retry := input.Retry != nil && *input.Retry

		// Enqueue email job (will be skipped by deduplication logic)
		if enqueueEmailJob != nil {
			_ = enqueueEmailJob(s.DB, input.Email, existingOTP.Code, retry, ip, browser, os)
		}

		return &model.CreateOtpResponse{
			Code:    200,
			Success: true,
			Message: "Verification code has been sent to your email",
		}, nil
	}

	// Generate 6-digit OTP code
	code := s.generateOTPCode()

	// Invalidate any existing OTPs for this email and type
	s.DB.Model(&model.OtpDB{}).Where("email = ? AND type = ?", input.Email, input.Type).Update("is_valid", false)

	// Create new OTP record
	otpDB := model.OtpDB{
		Email:     input.Email,
		Code:      code,
		IsValid:   true,
		ExpiredAt: time.Now().Add(10 * time.Minute),
		Type:      input.Type,
	}

	result = s.DB.Create(&otpDB)
	if result.Error != nil {
		return &model.CreateOtpResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to create OTP: %v", result.Error),
		}, nil
	}

	var ip, browser, os *string
	if clientInfo != nil {
		ip = &clientInfo.IP
		browser = &clientInfo.Browser
		os = &clientInfo.OS
	}

	// Determine retry value
	retry := input.Retry != nil && *input.Retry

	// Enqueue email job to background queue
	if enqueueEmailJob != nil {
		err := enqueueEmailJob(s.DB, input.Email, code, retry, ip, browser, os)
		if err != nil {
			// Log the error but still return success (OTP is saved in DB)
			fmt.Printf("Warning: Failed to enqueue email job: %v\n", err)
		}
	}

	return &model.CreateOtpResponse{
		Code:    200,
		Success: true,
		Message: "Verification code has been sent to your email",
	}, nil
}

// VerifyOtp verifies an OTP code and returns a token for password reset
func (s *AuthService) VerifyOtp(ctx context.Context, input model.VerifyOtpInput) (*model.VerifyOtpResponse, error) {
	// Find the OTP record
	var otpDB model.OtpDB
	result := s.DB.Where("email = ? AND code = ? AND type = ? AND is_valid = ?", input.Email, input.Code, input.Type, true).First(&otpDB)
	if result.Error != nil {
		return &model.VerifyOtpResponse{
			Code:    400,
			Success: false,
			Message: "Invalid or expired verification code",
			Token:   nil,
		}, nil
	}

	// Check if OTP is expired
	if time.Now().After(otpDB.ExpiredAt) {
		// Mark as invalid
		s.DB.Model(&otpDB).Update("is_valid", false)
		return &model.VerifyOtpResponse{
			Code:    400,
			Success: false,
			Message: "Verification code has expired",
			Token:   nil,
		}, nil
	}

	// Find the user
	var userDB model.UserDB
	result = s.DB.Where("email = ? AND is_active = ?", input.Email, true).First(&userDB)
	if result.Error != nil {
		return &model.VerifyOtpResponse{
			Code:    404,
			Success: false,
			Message: "User not found",
			Token:   nil,
		}, nil
	}

	// Get user role
	var userRole model.UserRoleDB
	roleName := ""
	if userDB.RoleID != nil {
		s.DB.First(&userRole, *userDB.RoleID)
		if userRole.ID > 0 {
			roleName = userRole.Name
		}
	}

	// Generate JWT token for password reset
	secureID := ""
	if userDB.SecureID != nil {
		secureID = *userDB.SecureID
	}

	token, err := s.generateJWT(userDB.ID, userDB.Email, roleName, secureID, "password_reset")
	if err != nil {
		return &model.VerifyOtpResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to generate token: %v", err),
			Token:   nil,
		}, nil
	}

	// Invalidate the OTP (one-time use)
	s.DB.Model(&otpDB).Update("is_valid", false)

	return &model.VerifyOtpResponse{
		Code:    200,
		Success: true,
		Message: "Verification successful",
		Token:   &token,
	}, nil
}

// NewPassword updates the user's password
func (s *AuthService) NewPassword(ctx context.Context, input model.NewPasswordInput, userClaims *Claims) (*model.NewPasswordResponse, error) {
	if userClaims == nil {
		return &model.NewPasswordResponse{
			Code:    401,
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	// Verify that the token purpose is password_reset (extra safety check)
	if userClaims.Purpose != "password_reset" {
		return &model.NewPasswordResponse{
			Code:    403,
			Success: false,
			Message: "Token is not authorized for password reset",
		}, nil
	}

	// Find the user by ID
	var userDB model.UserDB
	result := s.DB.Where("id = ? AND is_active = ?", userClaims.UserID, true).First(&userDB)
	if result.Error != nil {
		return &model.NewPasswordResponse{
			Code:    404,
			Success: false,
			Message: "User not found",
		}, nil
	}

	// Hash the new password
	hashedPassword, err := helper.HashPassword(input.Password)
	if err != nil {
		return &model.NewPasswordResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to hash password: %v", err),
		}, nil
	}

	// Update the user's password
	userDB.Password = hashedPassword
	result = s.DB.Save(&userDB)
	if result.Error != nil {
		return &model.NewPasswordResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to update password: %v", result.Error),
		}, nil
	}

	return &model.NewPasswordResponse{
		Code:    200,
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

// Helper methods

func (s *AuthService) getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func (s *AuthService) getJWTSecret() []byte {
	return []byte(s.getEnv("JWT_SECRET", "your-256-bit-secret"))
}

func (s *AuthService) generateOTPCode() string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < 6; i++ {
		code += string(rune('0' + rand.Intn(10)))
	}
	return code
}

func (s *AuthService) generateJWT(userID int32, email string, role string, secureID string, purpose string) (string, error) {
	jwtSecret := s.getJWTSecret()
	expiry := 30 * 24 * time.Hour // 30 days for login tokens (actual expiry controlled by database)
	if purpose == "password_reset" {
		expiry = 15 * time.Minute
	}
	expirationTime := time.Now().Add(expiry)

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		SecureID: secureID,
		Purpose:  purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kasirpinter",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) validateJWT(tokenString string) (*Claims, error) {
	jwtSecret := s.getJWTSecret()
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *AuthService) blacklistToken(tokenString string, userID int32, expiresAt time.Time) error {
	blacklistedToken := model.BlacklistedTokenDB{
		Token:     tokenString,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	return s.DB.Create(&blacklistedToken).Error
}

func (s *AuthService) cleanupExpiredBlacklistedTokens() error {
	return s.DB.Where("expires_at <= ?", time.Now()).Delete(&model.BlacklistedTokenDB{}).Error
}

// RefreshToken generates a new access token using a valid access token
func (s *AuthService) RefreshToken(ctx context.Context, input input.RefreshTokenInput) (*model.RefreshTokenResponse, error) {
	// Validate the access token
	claims, err := s.validateJWT(input.Token)
	if err != nil {
		return &model.RefreshTokenResponse{
			Code:    401,
			Success: false,
			Message: "Invalid token",
		}, nil
	}

	// Check if the token is blacklisted
	if s.isTokenBlacklisted(input.Token) {
		return &model.RefreshTokenResponse{
			Code:    401,
			Success: false,
			Message: "Token has been revoked",
		}, nil
	}

	// Find the user
	var userDB model.UserDB
	result := s.DB.Where("id = ? AND is_active = ?", claims.UserID, true).First(&userDB)
	if result.Error != nil {
		return &model.RefreshTokenResponse{
			Code:    404,
			Success: false,
			Message: "User not found",
		}, nil
	}

	// Get user role
	var userRole model.UserRoleDB
	roleName := ""
	if userDB.RoleID != nil {
		s.DB.First(&userRole, *userDB.RoleID)
		if userRole.ID > 0 {
			roleName = userRole.Name
		}
	}

	secureID := ""
	if userDB.SecureID != nil {
		secureID = *userDB.SecureID
	}

	// Generate new access token
	newToken, err := s.generateJWT(userDB.ID, userDB.Email, roleName, secureID, "login")
	if err != nil {
		return &model.RefreshTokenResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to generate new token: %v", err),
		}, nil
	}

	// Blacklist the old token
	err = s.blacklistToken(input.Token, claims.UserID, claims.ExpiresAt.Time)
	if err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to blacklist old token: %v\n", err)
	}

	// Convert DB model to GraphQL model using mapper
	var userRoleDB *model.UserRoleDB
	if userRole.ID > 0 {
		userRoleDB = &userRole
	}
	user := helper.ToGraphQLUser(userDB, userRoleDB)

	return &model.RefreshTokenResponse{
		Code:    200,
		Success: true,
		Message: "Token refreshed successfully",
		Data: &model.AuthData{
			Token: newToken,
			User:  user,
		},
	}, nil
}

// LogoutExpiredToken handles logout for expired tokens by blacklisting them
func (s *AuthService) LogoutExpiredToken(ctx context.Context, token string) (*model.LogoutResponse, error) {
	// Check if token is already blacklisted
	if s.isTokenBlacklisted(token) {
		return &model.LogoutResponse{
			Code:    200,
			Success: true,
			Message: "Token already blacklisted",
		}, nil
	}

	// Try to extract user info from the expired token for logging
	claims, err := s.validateJWT(token)
	var userID int32
	if err == nil {
		userID = claims.UserID
	} else {
		// Token is invalid/expired, but we'll still blacklist it
		userID = 0
	}

	// Blacklist the expired token
	// Set expiry to far future to ensure it stays blacklisted
	farFuture := time.Now().Add(365 * 24 * time.Hour) // 1 year
	err = s.blacklistToken(token, userID, farFuture)
	if err != nil {
		return &model.LogoutResponse{
			Code:    500,
			Success: false,
			Message: fmt.Sprintf("failed to blacklist token: %v", err),
		}, nil
	}

	// Also remove from active_tokens if it exists
	s.DB.Where("token = ?", token).Delete(&model.ActiveTokenDB{})

	return &model.LogoutResponse{
		Code:    200,
		Success: true,
		Message: "Expired token blacklisted successfully",
	}, nil
}

// isTokenBlacklisted checks if a token is in the blacklist
func (s *AuthService) isTokenBlacklisted(tokenString string) bool {
	var count int64
	s.DB.Model(&model.BlacklistedTokenDB{}).Where("token = ? AND expires_at > ?", tokenString, time.Now()).Count(&count)
	return count > 0
}
