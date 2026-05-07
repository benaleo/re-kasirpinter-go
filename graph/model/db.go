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

// ActiveTokenDB represents the database model for active tokens with managed expiry
type ActiveTokenDB struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null;size:500" json:"token"`
	UserID    int32     `gorm:"not null;index" json:"user_id"`
	IP        *string   `json:"ip,omitempty"`
	Browser   *string   `json:"browser,omitempty"`
	OS        *string   `json:"os,omitempty"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for ActiveTokenDB
func (ActiveTokenDB) TableName() string {
	return "active_tokens"
}

// BeforeCreate hook for ActiveTokenDB
func (a *ActiveTokenDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for ActiveTokenDB
func (a *ActiveTokenDB) BeforeUpdate(tx *gorm.DB) error {
	a.UpdatedAt = time.Now()
	return nil
}

// BlacklistedTokenDB represents the database model for blacklisted tokens
type BlacklistedTokenDB struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null;size:500" json:"token"`
	UserID    int32     `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
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

// IngredientCategoryDB represents the database model for IngredientCategory
type IngredientCategoryDB struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	Unit        string     `gorm:"not null" json:"unit"`
	ConvertUnit *string    `json:"convert_unit,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName specifies the table name for IngredientCategoryDB
func (IngredientCategoryDB) TableName() string {
	return "ingredient_categories"
}

// BeforeCreate hook for IngredientCategoryDB
func (i *IngredientCategoryDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	i.CreatedAt = now
	i.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for IngredientCategoryDB
func (i *IngredientCategoryDB) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}

// IngredientDB represents the database model for Ingredient
type IngredientDB struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string     `gorm:"not null" json:"name"`
	CategoryID *int64     `json:"category_id,omitempty"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Relations
	Category *IngredientCategoryDB `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Stocks   []IngredientStockDB   `gorm:"foreignKey:IngredientID" json:"stocks,omitempty"`
}

// TableName specifies the table name for IngredientDB
func (IngredientDB) TableName() string {
	return "ingredients"
}

// BeforeCreate hook for IngredientDB
func (i *IngredientDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	i.CreatedAt = now
	i.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for IngredientDB
func (i *IngredientDB) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}

// IngredientStockType represents the type of stock movement
type IngredientStockType string

const (
	IngredientStockTypeIncrease IngredientStockType = "increase"
	IngredientStockTypeDecrease IngredientStockType = "decrease"
)

// IngredientStockDB represents the database model for IngredientStock
type IngredientStockDB struct {
	ID          int64               `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        *string             `json:"code,omitempty"`
	Qty         float64             `gorm:"default:0" json:"qty"`
	Type        IngredientStockType `gorm:"not null" json:"type"`
	Capital     float64             `json:"capital"`
	CapitalItem float64             `json:"capital_item"`
	Message     *string             `json:"message,omitempty"`
	Image       *string             `json:"image,omitempty"`
	DeletedAt   *time.Time          `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`

	// Relations
	IngredientID int64         `gorm:"not null;index" json:"ingredient_id"`
	Ingredient   *IngredientDB `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// TableName specifies the table name for IngredientStockDB
func (IngredientStockDB) TableName() string {
	return "ingredient_stocks"
}

// BeforeCreate hook for IngredientStockDB
func (i *IngredientStockDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	i.CreatedAt = now
	i.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for IngredientStockDB
func (i *IngredientStockDB) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}

// ProductCategoryDB represents the database model for ProductCategory
type ProductCategoryDB struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	Parent   *ProductCategoryDB  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []ProductCategoryDB `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Products []ProductDB         `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// TableName specifies the table name for ProductCategoryDB
func (ProductCategoryDB) TableName() string {
	return "product_categories"
}

// BeforeCreate hook for ProductCategoryDB
func (p *ProductCategoryDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for ProductCategoryDB
func (p *ProductCategoryDB) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// ProductDB represents the database model for Product
type ProductDB struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	SecureID      *string    `gorm:"uniqueIndex" json:"secure_id,omitempty"`
	Name          string     `gorm:"not null" json:"name"`
	Image         *string    `json:"image,omitempty"`
	CategoryID    *int64     `json:"category_id,omitempty"`
	Description   *string    `json:"description,omitempty"`
	AvailableType *string    `json:"available_type,omitempty"`
	VariantType   *string    `json:"variant_type,omitempty"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relations
	Category         *ProductCategoryDB  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Variants         []ProductVariantDB  `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	ProductHasExtras []ProductHasExtraDB `gorm:"foreignKey:ProductID" json:"product_has_extras,omitempty"`
}

// TableName specifies the table name for ProductDB
func (ProductDB) TableName() string {
	return "products"
}

// BeforeCreate hook for ProductDB
func (p *ProductDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for ProductDB
func (p *ProductDB) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercent DiscountType = "percent"
	DiscountTypeAmount  DiscountType = "amount"
)

// DiscountDB represents the database model for Discount
type DiscountDB struct {
	ID          int64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string       `gorm:"not null" json:"name"`
	Description *string      `json:"description,omitempty"`
	Icon        *string      `json:"icon,omitempty"`
	Code        *string      `gorm:"uniqueIndex" json:"code,omitempty"`
	Type        DiscountType `gorm:"not null" json:"type"`
	Value       float64      `gorm:"not null" json:"value"`
	MaxValue    *int32       `json:"max_value,omitempty"`
	MinOrder    *int32       `json:"min_order,omitempty"`
	Quota       *int32       `json:"quota,omitempty"`
	StartAt     *time.Time   `json:"start_at,omitempty"`
	EndAt       *time.Time   `json:"end_at,omitempty"`
	IsActive    bool         `gorm:"default:true" json:"is_active"`
	DeletedAt   *time.Time   `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// TableName specifies the table name for DiscountDB
func (DiscountDB) TableName() string {
	return "discounts"
}

// BeforeCreate hook for DiscountDB
func (d *DiscountDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	d.CreatedAt = now
	d.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for DiscountDB
func (d *DiscountDB) BeforeUpdate(tx *gorm.DB) error {
	d.UpdatedAt = time.Now()
	return nil
}

// ProductVariantDB represents the database model for ProductVariant
type ProductVariantDB struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Image         *string    `json:"image,omitempty"`
	ProductID     int64      `gorm:"not null;index" json:"product_id"`
	Name          string     `gorm:"not null" json:"name"`
	Price         float64    `gorm:"not null" json:"price"`
	PriceOriginal *float64   `json:"price_original,omitempty"`
	IsUnlimited   bool       `gorm:"default:true" json:"is_unlimited"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relations
	Product     *ProductDB            `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Ingredients []ProductIngredientDB `gorm:"foreignKey:VariantID" json:"ingredients,omitempty"`
}

// TableName specifies the table name for ProductVariantDB
func (ProductVariantDB) TableName() string {
	return "product_variants"
}

// BeforeCreate hook for ProductVariantDB
func (p *ProductVariantDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for ProductVariantDB
func (p *ProductVariantDB) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// ProductIngredientDB represents the database model for ProductIngredient
type ProductIngredientDB struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	VariantID       int64     `gorm:"not null;index" json:"variant_id"`
	IngredientID    int64     `gorm:"not null;index" json:"ingredient_id"`
	IngredientValue float64   `gorm:"not null" json:"ingredient_value"`
	Unit            string    `gorm:"not null" json:"unit"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relations
	Variant    *ProductVariantDB `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
	Ingredient *IngredientDB     `gorm:"foreignKey:IngredientID" json:"ingredient,omitempty"`
}

// TableName specifies the table name for ProductIngredientDB with unique constraint
func (ProductIngredientDB) TableName() string {
	return "product_ingredients"
}

// AddUniqueIndexes adds the unique constraint for variant_id and ingredient_id
func (ProductIngredientDB) AddUniqueIndexes(db *gorm.DB) error {
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_variant_ingredient ON product_ingredients(variant_id, ingredient_id)").Error
}

// BeforeCreate hook for ProductIngredientDB
func (p *ProductIngredientDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// ProductExtraDB represents the database model for ProductExtra
type ProductExtraDB struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"not null" json:"name"`
	Price     float64    `gorm:"not null" json:"price"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName specifies the table name for ProductExtraDB
func (ProductExtraDB) TableName() string {
	return "product_extras"
}

// BeforeCreate hook for ProductExtraDB
func (p *ProductExtraDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for ProductExtraDB
func (p *ProductExtraDB) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// ProductHasExtraDB represents the database model for ProductHasExtra
type ProductHasExtraDB struct {
	ProductID      int64     `gorm:"primaryKey;not null" json:"product_id"`
	ProductExtraID int64     `gorm:"primaryKey;not null" json:"product_extra_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Product      *ProductDB      `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductExtra *ProductExtraDB `gorm:"foreignKey:ProductExtraID" json:"product_extra,omitempty"`
}

// TableName specifies the table name for ProductHasExtraDB
func (ProductHasExtraDB) TableName() string {
	return "product_has_extras"
}

// AddUniqueIndexes adds the unique constraint for product_id and product_extra_id
func (ProductHasExtraDB) AddUniqueIndexes(db *gorm.DB) error {
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_product_extra ON product_has_extras(product_id, product_extra_id)").Error
}

// BeforeCreate hook for ProductHasExtraDB
func (p *ProductHasExtraDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// TransactionDB represents the database model for Transaction
type TransactionDB struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	SecureID      *string   `gorm:"uniqueIndex;index" json:"secure_id,omitempty"`
	Date          string    `gorm:"type:date;not null;index" json:"date"`
	Sequence      int32     `gorm:"not null" json:"sequence"`
	Invoice       string    `gorm:"uniqueIndex;not null" json:"invoice"`
	PaymentMethod string    `gorm:"not null" json:"payment_method"`
	TotalAmount   float64   `gorm:"not null" json:"total_amount"`
	TotalBilled   float64   `gorm:"not null" json:"total_billed"`
	Tax           float64   `gorm:"default:0" json:"tax"`
	Subtotal      float64   `gorm:"not null" json:"subtotal"`
	Discount      float64   `gorm:"default:0" json:"discount"`
	CustomerID    *string   `gorm:"index" json:"customer_id,omitempty"`
	IsCompleted   bool      `gorm:"default:false" json:"is_completed"`
	IsCanceled    bool      `gorm:"default:false" json:"is_canceled"`
	CreatedAt     time.Time `json:"created_at"`
	CreatedBy     *string   `json:"created_by,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
	UpdatedBy     *string   `json:"updated_by,omitempty"`

	// Relations
	Customer *UserDB                `gorm:"foreignKey:CustomerID;references:SecureID" json:"customer,omitempty"`
	Products []TransactionProductDB `gorm:"foreignKey:TransactionID" json:"products,omitempty"`
}

// TableName specifies the table name for TransactionDB
func (TransactionDB) TableName() string {
	return "transactions"
}

// BeforeCreate hook for TransactionDB
func (t *TransactionDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

// BeforeUpdate hook for TransactionDB
func (t *TransactionDB) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

// TransactionProductDB represents the database model for TransactionProduct
type TransactionProductDB struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TransactionID int64     `gorm:"not null;index" json:"transaction_id"`
	ProductID     int64     `gorm:"not null;index" json:"product_id"`
	AvailableType string    `gorm:"not null" json:"available_type"`
	VariantType   string    `gorm:"not null" json:"variant_type"`
	Attribute     string    `json:"attribute,omitempty"`
	VariantName   string    `gorm:"not null" json:"variant_name"`
	Quantity      int32     `gorm:"not null" json:"quantity"`
	ProductPrice  float64   `gorm:"not null" json:"product_price"`
	TotalExtras   float64   `gorm:"default:0" json:"total_extras"`
	TotalPrice    float64   `gorm:"not null" json:"total_price"`
	Notes         *string   `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	// Relations
	Transaction *TransactionDB       `gorm:"foreignKey:TransactionID" json:"transaction,omitempty"`
	Product     *ProductDB           `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Extras      []TransactionExtraDB `gorm:"foreignKey:TransactionProductID" json:"extras,omitempty"`
}

// TableName specifies the table name for TransactionProductDB
func (TransactionProductDB) TableName() string {
	return "transaction_products"
}

// BeforeCreate hook for TransactionProductDB
func (t *TransactionProductDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.CreatedAt = now
	return nil
}

// TransactionExtraDB represents the database model for TransactionExtra
type TransactionExtraDB struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TransactionProductID int64     `gorm:"not null;index" json:"transaction_product_id"`
	ExtraName            string    `gorm:"not null" json:"extra_name"`
	Quantity             int32     `gorm:"not null" json:"quantity"`
	ExtraPrice           float64   `gorm:"not null" json:"extra_price"`
	CreatedAt            time.Time `json:"created_at"`

	// Relations
	TransactionProduct *TransactionProductDB `gorm:"foreignKey:TransactionProductID" json:"transaction_product,omitempty"`
}

// TableName specifies the table name for TransactionExtraDB
func (TransactionExtraDB) TableName() string {
	return "transaction_extras"
}

// BeforeCreate hook for TransactionExtraDB
func (t *TransactionExtraDB) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.CreatedAt = now
	return nil
}
