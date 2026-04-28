package utils

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

type PaginationMetadata struct {
	Page       int32  `json:"page"`
	Limit      int32  `json:"limit"`
	Total      int32  `json:"total"`
	TotalPages int32  `json:"total_pages"`
	Links      Links  `json:"links"`
}

type Links struct {
	Self string  `json:"self"`
	Next *string `json:"next"`
	Prev *string `json:"prev"`
}

// Response represents a standard API response structure
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// JSONResponse sends a successful JSON response
func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Status: "success",
		Data:   data,
	})
}

// JSONResponseWithMessage sends a success response with a custom message and data
func JSONResponseWithMessage(w http.ResponseWriter, status int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

// PaginatedResponse sends a standardized paginated response
func PaginatedResponse(w http.ResponseWriter, status int, page, limit, total int32, path string, query url.Values, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	totalPages := (total + limit - 1) / limit

	// Helper to build links
	buildLink := func(p int32) string {
		q := url.Values{}
		for k, v := range query {
			q[k] = v
		}
		q.Set("page", strconv.Itoa(int(p)))
		q.Set("limit", strconv.Itoa(int(limit)))
		return path + "?" + q.Encode()
	}

	links := Links{
		Self: buildLink(page),
	}

	if page < totalPages {
		next := buildLink(page + 1)
		links.Next = &next
	}
	if page > 1 {
		prev := buildLink(page - 1)
		links.Prev = &prev
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status":      "success",
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
		"links":       links,
		"data":        data,
	})
}

// ErrorResponse sends an error JSON response
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Status:  "error",
		Message: message,
	})
}
