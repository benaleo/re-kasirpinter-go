# Project Structure

This document describes the structure of the **re-kasirpinter-go** project, a Go-based GraphQL API for a point-of-sale (POS) system.

## Overview

This is a Go web application using GraphQL API with PostgreSQL database, JWT authentication, and cloud storage integration.

## Directory Structure

```
re-kasirpinter-go/
├── cmd/                          # Command-line utilities
│   ├── README.md                 # Documentation for CLI commands
│   └── migrate_fresh.go          # Database migration utility
├── config/                       # Configuration management
│   ├── db.go                     # Database configuration and connection
│   └── env.go                    # Environment variable handling
├── graph/                        # GraphQL implementation
│   ├── auth.go                   # Authentication logic
│   ├── email.go                  # Email functionality
│   ├── email_queue.go            # Email queue management
│   ├── generated.go              # Auto-generated GraphQL code
│   ├── mappers.go               # Data mapping utilities
│   ├── middleware.go            # GraphQL middleware
│   ├── resolver.go              # Main GraphQL resolver
│   ├── schema.graphqls          # GraphQL schema definition
│   ├── schema.resolvers.go      # GraphQL resolvers implementation
│   ├── input/                   # GraphQL input types
│   │   ├── auth_login_input.go
│   │   ├── input.graphqls
│   │   ├── user_create_input.go
│   │   └── user_update_input.go
│   └── model/                   # Data models
│       ├── auth_models.go
│       ├── auth_types.go
│       ├── db.go                # Database models
│       ├── model.graphqls
│       └── models_gen.go        # Auto-generated models
├── helper/                      # Utility functions
│   ├── aes.go                   # AES encryption utilities
│   ├── auth.go                  # Authentication helpers
│   ├── mapper.go                # Data mapping helpers
│   ├── pagination.go            # Pagination utilities
│   └── response.go              # Response formatting
├── service/                     # Business logic services
│   ├── auth_service.go          # Authentication service
│   ├── cloudflare_r2_service.go # Cloud storage service
│   ├── ingredient_category_service.go
│   ├── ingredient_service.go
│   ├── ingredient_stock_service.go
│   ├── product_category_service.go
│   ├── product_service.go
│   ├── role_service.go
│   └── user_service.go
├── templates/                   # Template files
├── .dockerignore               # Docker ignore file
├── .env                        # Environment variables (local)
├── .env.example               # Environment variables template
├── .gitignore                  # Git ignore file
├── DEPLOY.md                   # Deployment documentation
├── Dockerfile                  # Docker container configuration
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── gqlgen.yml                  # GraphQL code generation config
├── koyeb.yml                   # Koyeb deployment config
├── leapcell.json              # Leapcell deployment config
├── leapcell.yaml              # Leapcell deployment config
├── render.yaml                # Render deployment config
└── server.go                  # Main application entry point
```

## Key Components

### Core Application
- **server.go**: Main entry point with HTTP server setup and CORS middleware
- **config/**: Database and environment configuration management
- **service/**: Business logic layer handling core operations

### GraphQL API
- **graph/**: Complete GraphQL implementation using gqlgen
- **schema.graphqls**: API schema definition
- **resolver.go**: Main resolver interface
- **schema.resolvers.go**: Implementation of all GraphQL resolvers

### Data Models
- **graph/model/**: Database models and GraphQL types
- **helper/mapper.go**: Data transformation between layers

### Utilities
- **helper/**: Shared utilities for encryption, pagination, and response formatting
- **cmd/**: CLI tools for database operations

## Technology Stack

- **Language**: Go 1.25.6
- **GraphQL**: gqlgen framework
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT tokens
- **Storage**: Cloudflare R2 (S3-compatible)
- **Email**: gomail library
- **Deployment**: Docker containers with multiple platform support

## Key Features

- User authentication and authorization
- Product and ingredient management
- Stock tracking
- Role-based access control
- Email notifications
- File upload to cloud storage
- RESTful GraphQL API
- Database migrations

## Configuration

The application uses environment variables for configuration (see `.env.example`):
- Database connection
- JWT secrets
- Cloud storage credentials
- Email settings
- CORS origins

## Deployment

Supported deployment platforms:
- Docker
- Koyeb
- Leapcell
- Render

See `DEPLOY.md` for detailed deployment instructions.

## Adding GraphQL APIs - Complete Flow

### Step 1: Define GraphQL Schema
**File**: `@graph/schema.graphqls`

Add your new types, queries, and mutations:

```graphql
# Example: Adding a new Entity
type NewEntity {
  id: Int64!
  name: String!
  description: String
  is_active: Boolean!
  created_at: Time!
  updated_at: Time!
}

type Query {
  # Add new query
  newEntities(pagination: PaginationInput): NewEntitiesResponse! @auth
  newEntity(id: Int64!): NewEntityResponse! @auth
}

type Mutation {
  # Add new mutations
  createNewEntity(input: CreateNewEntityInput!): CreateNewEntityResponse! @auth
  updateNewEntity(id: Int64!, input: UpdateNewEntityInput!): UpdateNewEntityResponse! @auth
  deleteNewEntity(id: Int64!): DeleteNewEntityResponse! @auth
}
```

### Step 2: Define Input Types
**File**: `@graph/input/input.graphqls`

Add input types for your mutations:

```graphql
input CreateNewEntityInput {
  name: String!
  description: String
  is_active: Boolean!
}

input UpdateNewEntityInput {
  name: String
  description: String
  is_active: Boolean
}
```

### Step 3: Define Response Types
**File**: `@graph/model/model.graphqls`

Add response types for your API:

```graphql
type NewEntityResponse {
  code: Int!
  success: Boolean!
  message: String!
  data: NewEntity
}

type NewEntitiesResponse {
  code: Int!
  success: Boolean!
  message: String!
  data: [NewEntity!]!
  pagination: PageInfo!
}

type CreateNewEntityResponse {
  code: Int!
  success: Boolean!
  message: String!
  data: NewEntity
}

type UpdateNewEntityResponse {
  code: Int!
  success: Boolean!
  message: String!
  data: NewEntity
}

type DeleteNewEntityResponse {
  code: Int!
  success: Boolean!
  message: String!
  data: NewEntity
}
```

### Step 4: Generate GraphQL Code
**Command**: `go tool gqlgen generate`

This will:
- Generate resolver stubs in `@graph/schema.resolvers.go`
- Update `@graph/generated.go` with new types
- Generate model files in `@graph/model/models_gen.go`

### Step 5: Create Mapper Functions
**File**: `@graph/mappers.go`

Add GraphQL response mappers:

```go
// Response mappers for NewEntity
func toGraphQLCreateNewEntityResponse(code int32, success bool, message string, data *model.NewEntity) *model.CreateNewEntityResponse {
    return &model.CreateNewEntityResponse{
        Code:    code,
        Success: success,
        Message: message,
        Data:    data,
    }
}

func toGraphQLUpdateNewEntityResponse(code int32, success bool, message string, data *model.NewEntity) *model.UpdateNewEntityResponse {
    return &model.UpdateNewEntityResponse{
        Code:    code,
        Success: success,
        Message: message,
        Data:    data,
    }
}

func toGraphQLDeleteNewEntityResponse(code int32, success bool, message string, data *model.NewEntity) *model.DeleteNewEntityResponse {
    return &model.DeleteNewEntityResponse{
        Code:    code,
        Success: success,
        Message: message,
        Data:    data,
    }
}

func toGraphQLNewEntitiesResponse(code int32, success bool, message string, data []*model.NewEntity, pagination *model.PageInfo) *model.NewEntitiesResponse {
    return &model.NewEntitiesResponse{
        Code:       code,
        Success:    success,
        Message:    message,
        Data:       data,
        Pagination: pagination,
    }
}

func toGraphQLNewEntityResponse(code int32, success bool, message string, data *model.NewEntity) *model.NewEntityResponse {
    return &model.NewEntityResponse{
        Code:    code,
        Success: success,
        Message: message,
        Data:    data,
    }
}
```

**File**: `@helper/mapper.go`

Add database model to GraphQL model converters:

```go
// ToGraphQLNewEntity converts NewEntityDB to GraphQL NewEntity model
func ToGraphQLNewEntity(newEntityDB model.NewEntityDB) *model.NewEntity {
    return &model.NewEntity{
        ID:          newEntityDB.ID,
        Name:        newEntityDB.Name,
        Description: newEntityDB.Description,
        IsActive:    newEntityDB.IsActive,
        DeletedAt:   newEntityDB.DeletedAt,
        CreatedAt:   newEntityDB.CreatedAt,
        UpdatedAt:   newEntityDB.UpdatedAt,
    }
}

// ToGraphQLNewEntitySlice converts []NewEntityDB to []*model.NewEntity
func ToGraphQLNewEntitySlice(newEntitiesDB []model.NewEntityDB) []*model.NewEntity {
    entities := make([]*model.NewEntity, len(newEntitiesDB))
    for i, entityDB := range newEntitiesDB {
        entities[i] = ToGraphQLNewEntity(entityDB)
    }
    return entities
}
```

### Step 6: Implement Resolvers
**File**: `@graph/schema.resolvers.go`

Implement the resolver methods using the mappers:

```go
// CreateNewEntity is the resolver for the createNewEntity field.
func (r *mutationResolver) CreateNewEntity(ctx context.Context, input input.CreateNewEntityInput) (*model.CreateNewEntityResponse, error) {
    if r.NewEntityService == nil {
        return toGraphQLCreateNewEntityResponse(500, false, "service not initialized", nil), nil
    }
    
    entityDB, err := r.NewEntityService.Create(ctx, input)
    if err != nil {
        return toGraphQLCreateNewEntityResponse(400, false, err.Error(), nil), nil
    }
    
    entity := helper.ToGraphQLNewEntity(*entityDB)
    return toGraphQLCreateNewEntityResponse(200, true, "New entity created successfully", entity), nil
}

// UpdateNewEntity is the resolver for the updateNewEntity field.
func (r *mutationResolver) UpdateNewEntity(ctx context.Context, id int64, input input.UpdateNewEntityInput) (*model.UpdateNewEntityResponse, error) {
    if r.NewEntityService == nil {
        return toGraphQLUpdateNewEntityResponse(500, false, "service not initialized", nil), nil
    }
    
    entityDB, err := r.NewEntityService.Update(ctx, id, input)
    if err != nil {
        return toGraphQLUpdateNewEntityResponse(400, false, err.Error(), nil), nil
    }
    
    entity := helper.ToGraphQLNewEntity(*entityDB)
    return toGraphQLUpdateNewEntityResponse(200, true, "New entity updated successfully", entity), nil
}

// DeleteNewEntity is the resolver for the deleteNewEntity field.
func (r *mutationResolver) DeleteNewEntity(ctx context.Context, id int64) (*model.DeleteNewEntityResponse, error) {
    if r.NewEntityService == nil {
        return toGraphQLDeleteNewEntityResponse(500, false, "service not initialized", nil), nil
    }
    
    entityDB, err := r.NewEntityService.Delete(ctx, id)
    if err != nil {
        return toGraphQLDeleteNewEntityResponse(400, false, err.Error(), nil), nil
    }
    
    entity := helper.ToGraphQLNewEntity(*entityDB)
    return toGraphQLDeleteNewEntityResponse(200, true, "New entity deleted successfully", entity), nil
}

// NewEntities is the resolver for the newEntities field.
func (r *queryResolver) NewEntities(ctx context.Context, pagination *model.PaginationInput) (*model.NewEntitiesResponse, error) {
    if r.NewEntityService == nil {
        return toGraphQLNewEntitiesResponse(500, false, "service not initialized", nil, nil), nil
    }
    
    entitiesDB, pageInfo, err := r.NewEntityService.GetAll(ctx, pagination)
    if err != nil {
        return toGraphQLNewEntitiesResponse(400, false, err.Error(), nil, nil), nil
    }
    
    entities := helper.ToGraphQLNewEntitySlice(entitiesDB)
    return toGraphQLNewEntitiesResponse(200, true, "New entities retrieved successfully", entities, pageInfo), nil
}

// NewEntity is the resolver for the newEntity field.
func (r *queryResolver) NewEntity(ctx context.Context, id int64) (*model.NewEntityResponse, error) {
    if r.NewEntityService == nil {
        return toGraphQLNewEntityResponse(500, false, "service not initialized", nil), nil
    }
    
    entityDB, err := r.NewEntityService.GetByID(ctx, id)
    if err != nil {
        return toGraphQLNewEntityResponse(400, false, err.Error(), nil), nil
    }
    
    entity := helper.ToGraphQLNewEntity(*entityDB)
    return toGraphQLNewEntityResponse(200, true, "New entity retrieved successfully", entity), nil
}
```

### Step 7: Add Database Model (if new entity)
**File**: `@graph/model/db.go`

Add the GORM model:

```go
type NewEntity struct {
    ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Name        string    `gorm:"not null" json:"name"`
    Description string    `json:"description"`
    IsActive    bool      `gorm:"default:true" json:"is_active"`
    DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func (NewEntity) TableName() string {
    return "new_entities"
}
```

### Step 8: Auto-migrate Database (if new entity)
**File**: `@config/db.go`

Add the new model to the auto-migration:

```go
// In the AutoMigrate function, add:
err = db.AutoMigrate(
    // ... existing models
    &graph.NewEntity{}, // Add this line
)
```

### Step 9: Create Service Layer (optional but recommended)
**File**: `@service/new_entity_service.go`

Create service for business logic:

```go
package service

type NewEntityService struct {
    DB *gorm.DB
}

func NewNewEntityService(db *gorm.DB) *NewEntityService {
    return &NewEntityService{DB: db}
}

func (s *NewEntityService) Create(ctx context.Context, input input.CreateNewEntityInput) (*graph.NewEntityDB, error) {
    // Business logic implementation
    entity := &graph.NewEntityDB{
        Name:        input.Name,
        Description: input.Description,
        IsActive:    input.IsActive,
    }
    
    if err := s.DB.Create(entity).Error; err != nil {
        return nil, err
    }
    
    return entity, nil
}

func (s *NewEntityService) Update(ctx context.Context, id int64, input input.UpdateNewEntityInput) (*graph.NewEntityDB, error) {
    var entity graph.NewEntityDB
    if err := s.DB.First(&entity, id).Error; err != nil {
        return nil, err
    }
    
    // Update fields if provided
    if input.Name != nil {
        entity.Name = *input.Name
    }
    if input.Description != nil {
        entity.Description = *input.Description
    }
    if input.IsActive != nil {
        entity.IsActive = *input.IsActive
    }
    
    if err := s.DB.Save(&entity).Error; err != nil {
        return nil, err
    }
    
    return &entity, nil
}

func (s *NewEntityService) Delete(ctx context.Context, id int64) (*graph.NewEntityDB, error) {
    var entity graph.NewEntityDB
    if err := s.DB.First(&entity, id).Error; err != nil {
        return nil, err
    }
    
    // Soft delete
    if err := s.DB.Delete(&entity).Error; err != nil {
        return nil, err
    }
    
    return &entity, nil
}

func (s *NewEntityService) GetAll(ctx context.Context, pagination *model.PaginationInput) ([]*graph.NewEntityDB, *model.PageInfo, error) {
    var entities []*graph.NewEntityDB
    var total int64
    
    // Build query
    query := s.DB.Where("deleted_at IS NULL")
    
    // Get total count
    if err := query.Model(&graph.NewEntityDB{}).Count(&total).Error; err != nil {
        return nil, nil, err
    }
    
    // Apply pagination
    limit := pagination.Limit
    offset := (pagination.Page - 1) * limit
    
    if err := query.Limit(limit).Offset(offset).Find(&entities).Error; err != nil {
        return nil, nil, err
    }
    
    // Build pagination info
    totalPages := int((total + int64(limit) - 1) / int64(limit))
    pageInfo := &model.PageInfo{
        CurrentPage:  pagination.Page,
        PerPage:      limit,
        TotalItems:   int(total),
        TotalPages:   totalPages,
        HasNextPage:  pagination.Page < totalPages,
        HasPreviousPage: pagination.Page > 1,
    }
    
    return entities, pageInfo, nil
}

func (s *NewEntityService) GetByID(ctx context.Context, id int64) (*graph.NewEntityDB, error) {
    var entity graph.NewEntityDB
    if err := s.DB.Where("id = ? AND deleted_at IS NULL", id).First(&entity).Error; err != nil {
        return nil, err
    }
    
    return &entity, nil
}
```

### Step 10: Register Service in Resolver
**File**: `@graph/resolver.go`

Add the new service to the resolver struct:

```go
type Resolver struct {
    // ... existing services
    NewEntityService *service.NewEntityService
}
```

And initialize it in `server.go` or wherever the resolver is created.

## Important Notes

1. **Authentication**: Add `@auth` directive to protected endpoints
2. **Validation**: Implement input validation in service layer
3. **Error Handling**: Use consistent response format with proper error codes
4. **Pagination**: Use `PaginationInput` for list queries
5. **Soft Deletes**: Use `deleted_at` for soft deletes (GORM convention)
6. **Naming**: Follow Go naming conventions and GraphQL schema conventions
7. **Testing**: Test each resolver method individually

## Example Complete Flow

For adding a complete new entity like "Category":
1. Add `Category` type to `@graph/schema.graphqls`
2. Add `CreateCategoryInput`, `UpdateCategoryInput` to `@graph/input/input.graphqls`
3. Add response types to `@graph/model/model.graphqls`
4. Run `go tool gqlgen generate`
5. **Add response mappers to `@graph/mappers.go`**
6. **Add model converters to `@helper/mapper.go`**
7. Implement resolvers in `@graph/schema.resolvers.go` (using mappers)
8. Add `Category` model to `@graph/model/db.go`
9. Add to auto-migration in `@config/db.go`
10. Create `@service/category_service.go` (using database models)
11. Register service in resolver
12. Test the implementation

## Mapper Pattern Summary

### `@graph/mappers.go` - Response Mappers
- Convert service results to GraphQL response types
- Follow naming: `toGraphQL[Action][Entity]Response`
- Handle consistent response structure (code, success, message, data)

### `@helper/mapper.go` - Model Converters  
- Convert database models to GraphQL types
- Follow naming: `ToGraphQL[Entity]` and `ToGraphQL[Entity]Slice`
- Handle data transformation and relationships

### Usage Flow
1. **Service Layer** → Returns database models (`*graph.NewEntityDB`)
2. **Helper Mapper** → Converts to GraphQL types (`*model.NewEntity`)
3. **Graph Mapper** → Wraps in response types (`*model.CreateNewEntityResponse`)
4. **Resolver** → Returns final response to client
