package helper

import (
	"re-kasirpinter-go/graph/model"
	"strings"
)

// PaginationParams holds the pagination configuration
type PaginationParams struct {
	Limit      int32
	Page       int32
	SortBy     string
	DefaultLimit int32
	DefaultPage  int32
	DefaultSortBy string
}

// PaginationResult holds the pagination metadata and query parameters
type PaginationResult struct {
	Limit      int32
	Page       int32
	Offset     int
	SortBy     string
	PageInfo   *model.PageInfo
}

// ParsePagination parses the pagination input and returns pagination parameters
func ParsePagination(pagination *model.PaginationInput) *PaginationParams {
	defaultLimit := int32(10)
	defaultPage := int32(1)
	defaultSortBy := "created_at,desc"

	if pagination != nil {
		if pagination.Limit != nil && *pagination.Limit > 0 {
			defaultLimit = *pagination.Limit
		}
		if pagination.Page != nil && *pagination.Page > 0 {
			defaultPage = *pagination.Page
		}
		if pagination.SortBy != nil && *pagination.SortBy != "" {
			defaultSortBy = *pagination.SortBy
		}
	}

	return &PaginationParams{
		Limit:      defaultLimit,
		Page:       defaultPage,
		SortBy:     defaultSortBy,
		DefaultLimit: 10,
		DefaultPage:  1,
		DefaultSortBy: "created_at,desc",
	}
}

// FormatSortBy converts "field,direction" to "field direction" for GORM
func FormatSortBy(sortBy string) string {
	if len(sortBy) == 0 {
		return sortBy
	}
	// Replace first comma with space
	if idx := strings.Index(sortBy, ","); idx != -1 {
		return sortBy[:idx] + " " + sortBy[idx+1:]
	}
	return sortBy
}

// BuildPaginationResult builds the pagination result with metadata
func BuildPaginationResult(params *PaginationParams, total int64, itemCount int) *PaginationResult {
	offset := (params.Page - 1) * params.Limit
	sortBy := FormatSortBy(params.SortBy)

	// Calculate pagination metadata
	totalPages := int32(0)
	if total > 0 {
		totalPages = (int32(total) + params.Limit - 1) / params.Limit
	}

	hasNextPage := params.Page < totalPages
	hasPreviousPage := params.Page > 1

	startItem := int32(0)
	endItem := int32(0)
	if total > 0 {
		startItem = offset + 1
		endItem = offset + int32(itemCount)
		if endItem > int32(total) {
			endItem = int32(total)
		}
	}

	// Build pagination info
	pageInfo := &model.PageInfo{
		CurrentPage:     params.Page,
		PerPage:         params.Limit,
		TotalItems:      int32(total),
		TotalPages:      totalPages,
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPreviousPage,
	}

	if total > 0 {
		pageInfo.StartItem = &startItem
		pageInfo.EndItem = &endItem
	}

	return &PaginationResult{
		Limit:    params.Limit,
		Page:     params.Page,
		Offset:   int(offset),
		SortBy:   sortBy,
		PageInfo: pageInfo,
	}
}
