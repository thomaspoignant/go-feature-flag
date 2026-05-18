package model

import "fmt"

type APIResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message,omitempty"`
	Data    any                 `json:"data,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalPages int `json:"totalPages"`
}

func NewPaginatedResponse[T any](items []T, total, page, pageSize int) PaginatedResponse[T] {
	totalPages := 0
	if total > 0 && pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return PaginatedResponse[T]{Data: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	return fmt.Sprintf("%d validation error(s)", len(v))
}

func (v ValidationErrors) ToMap() map[string][]string {
	m := make(map[string][]string)
	for _, e := range v {
		m[e.Field] = append(m[e.Field], e.Message)
	}
	return m
}
