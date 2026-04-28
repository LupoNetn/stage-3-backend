package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/luponetn/hng-stage-1/internals/db"
	"github.com/luponetn/hng-stage-1/utils"
)

// profileResponse sends a standard JSON response for profile data
func (h *Handler) profileResponse(w http.ResponseWriter, status int, msg string, p db.Profile) {
	data := map[string]any{
		"id":                  uuid.UUID(p.ID.Bytes).String(),
		"name":                p.Name,
		"gender":              p.Gender.String,
		"gender_probability":  p.GenderProbability.Float64,
		"country_name":        p.CountryName.String,
		"age":                 p.Age.Int32,
		"age_group":           p.AgeGroup.String,
		"country_id":          p.CountryID.String,
		"country_probability": p.CountryProbability.Float64,
		"created_at":          p.CreatedAt.Time.UTC().Format(time.RFC3339),
	}

	if msg == "success" {
		utils.JSONResponse(w, status, data)
	} else {
		utils.JSONResponseWithMessage(w, status, msg, data)
	}
}

// errorResponse delegates to the centralized ErrorResponse utility
func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	utils.ErrorResponse(w, status, message)
}

// mapToStruct converts a map to a struct using JSON marshaling
func mapToStruct(src any, dest any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

