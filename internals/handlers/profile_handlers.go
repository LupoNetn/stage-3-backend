package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/luponetn/hng-stage-1/internals/db"
	httprequest "github.com/luponetn/hng-stage-1/internals/httpRequest"
	"github.com/luponetn/hng-stage-1/utils"
)

func (h *Handler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		h.errorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Check for idempotency: Return existing if name already exists
	existingProfile, err := h.queries.GetProfileByName(r.Context(), name)
	if err == nil {
		h.profileResponse(w, http.StatusOK, "Profile already exists", existingProfile)
		return
	}

	// Call APIs concurrently using errgroup
	g, ctx := errgroup.WithContext(r.Context())

	var genderResp GenderizeResp
	var ageResp AgifyResp
	var countryResp NationalizeResp

	g.Go(func() error {
		url := fmt.Sprintf("https://api.genderize.io?name=%s", name)
		data, err := httprequest.MakeRequest(ctx, "GET", url, nil)
		if err != nil {
			return err
		}
		return mapToStruct(data, &genderResp)
	})

	g.Go(func() error {
		url := fmt.Sprintf("https://api.agify.io?name=%s", name)
		data, err := httprequest.MakeRequest(ctx, "GET", url, nil)
		if err != nil {
			return err
		}
		return mapToStruct(data, &ageResp)
	})

	g.Go(func() error {
		url := fmt.Sprintf("https://api.nationalize.io?name=%s", name)
		data, err := httprequest.MakeRequest(ctx, "GET", url, nil)
		if err != nil {
			return err
		}
		return mapToStruct(data, &countryResp)
	})

	if err := g.Wait(); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to fetch data from external APIs")
		return
	}

	// Edge Case Handling: Check for null/zero values from APIs
	if genderResp.Gender == "" || genderResp.Count == 0 {
		h.errorResponse(w, http.StatusBadGateway, "Predictive data for gender not available")
		return
	}

	if ageResp.Age == 0 { // Agify returns null/0 if age is unknown
		h.errorResponse(w, http.StatusBadGateway, "Predictive data for age not available")
		return
	}

	if len(countryResp.Country) == 0 {
		h.errorResponse(w, http.StatusBadGateway, "Predictive data for country not available")
		return
	}

	// Process country: Pick the one with highest probability
	var bestCountry string
	var bestProb float64
	for _, c := range countryResp.Country {
		if c.Probability > bestProb {
			bestProb = c.Probability
			bestCountry = c.CountryID
		}
	}

	// Process age group
	ageGroup := utils.ClassifyAgeGroup(ageResp.Age)

	// Create UUID v7
	id, err := uuid.NewV7()
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to generate ID")
		return
	}

	// Store in DB
	profile, err := h.queries.CreateProfile(r.Context(), db.CreateProfileParams{
		ID:                 utils.ToUUID(id),
		Name:               name,
		Gender:             utils.ToText(genderResp.Gender),
		GenderProbability:  utils.ToFloat8(genderResp.Probability),
		Age:                utils.ToInt4(ageResp.Age),
		AgeGroup:           utils.ToText(ageGroup),
		CountryID:          utils.ToText(bestCountry),
		CountryName:        utils.ToText(""), // nationalize.io does not provide full country names
		CountryProbability: utils.ToFloat8(bestProb),
	})

	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to create profile in database")
		return
	}

	h.profileResponse(w, http.StatusCreated, "success", profile)
}

func (h *Handler) GetProfileByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "Missing profile ID")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	profile, err := h.queries.GetProfile(r.Context(), utils.ToUUID(id))
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "Profile not found")
		return
	}

	h.profileResponse(w, http.StatusOK, "success", profile)
}

func (h *Handler) GetProfiles(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	page := params.Get("page")
	limit := params.Get("limit")
	gender := params.Get("gender")
	countryID := params.Get("country_id")
	ageGroup := params.Get("age_group")
	minAgeStr := params.Get("min_age")
	maxAgeStr := params.Get("max_age")
	minGenderProbStr := params.Get("min_gender_probability")
	minCountryProbStr := params.Get("min_country_probability")
	sortBy := strings.ToLower(params.Get("sort_by"))
	if sortBy == "" {
		sortBy = "created_at"
	} else if sortBy != "age" && sortBy != "created_at" && sortBy != "gender_probability" {
		h.errorResponse(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	sortDir := strings.ToLower(params.Get("order"))
	if sortDir == "" {
		sortDir = "desc"
	} else if sortDir != "asc" && sortDir != "desc" {
		h.errorResponse(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	limitVal := int32(10)
	if limit != "" {
		l, err := utils.ToInt32(limit)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		if l > 0 {
			limitVal = l
		}
	}
	if limitVal > 50 {
		limitVal = 50
	}

	pageVal := int32(1)
	if page != "" {
		p, err := utils.ToInt32(page)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		if p > 0 {
			pageVal = p
		}
	}

	offsetVal := (pageVal - 1) * limitVal

	// Prepare filters for SQL
	genders := []string{}
	if gender != "" {
		genders = []string{strings.ToLower(gender)}
	}

	minAge := int32(0)
	if minAgeStr != "" {
		v, err := utils.ToInt32(minAgeStr)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		minAge = v
	}
	maxAge := int32(0)
	if maxAgeStr != "" {
		v, err := utils.ToInt32(maxAgeStr)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		maxAge = v
	}
	minGenderProb := float64(0)
	if minGenderProbStr != "" {
		v, err := utils.ToFloat64(minGenderProbStr)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		minGenderProb = v
	}
	minCountryProb := float64(0)
	if minCountryProbStr != "" {
		v, err := utils.ToFloat64(minCountryProbStr)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		minCountryProb = v
	}

	countParams := db.CountProfilesAdvancedParams{
		Genders:        genders,
		AgeGroup:       strings.ToLower(ageGroup),
		CountryID:      strings.ToLower(countryID),
		MinAge:         minAge,
		MaxAge:         maxAge,
		MinGenderProb:  minGenderProb,
		MinCountryProb: minCountryProb,
	}

	total, err := h.queries.CountProfilesAdvanced(r.Context(), countParams)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to count profiles")
		return
	}

	listParams := db.ListProfilesAdvancedParams{
		Genders:        genders,
		AgeGroup:       strings.ToLower(ageGroup),
		CountryID:      strings.ToLower(countryID),
		MinAge:         minAge,
		MaxAge:         maxAge,
		MinGenderProb:  minGenderProb,
		MinCountryProb: minCountryProb,
		SortBy:         sortBy,
		SortDirection:  sortDir,
		LimitVal:       limitVal,
		OffsetVal:      offsetVal,
	}

	profiles, err := h.queries.ListProfilesAdvanced(r.Context(), listParams)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to fetch profiles")
		return
	}

	data := []map[string]any{}
	for _, p := range profiles {
		data = append(data, map[string]any{
			"id":                  uuid.UUID(p.ID.Bytes).String(),
			"name":                p.Name,
			"gender":              p.Gender.String,
			"gender_probability":  p.GenderProbability.Float64,
			"age":                 p.Age.Int32,
			"age_group":           p.AgeGroup.String,
			"country_id":          p.CountryID.String,
			"country_name":        p.CountryName.String,
			"country_probability": p.CountryProbability.Float64,
			"created_at":          p.CreatedAt.Time.UTC().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"page":   pageVal,
		"limit":  limitVal,
		"total":  total,
		"data":   data,
	})
}

func (h *Handler) SearchProfiles(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	q := params.Get("q")
	if q == "" {
		h.errorResponse(w, http.StatusBadRequest, "Missing query parameter")
		return
	}

	tokens := utils.NormalizeAndTokenize(q)

	filters, interpretable := utils.CreateRuleBasedFilter(tokens)
	if !interpretable {
		h.errorResponse(w, http.StatusUnprocessableEntity, "Unable to interpret query")
		return
	}

	// Pagination
	limitVal := int32(10)
	if lStr := params.Get("limit"); lStr != "" {
		l, err := utils.ToInt32(lStr)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		if l > 0 {
			limitVal = l
		}
	}
	if limitVal > 50 {
		limitVal = 50
	}
	pageVal := int32(1)
	if pStr := params.Get("page"); pStr != "" {
		p, err := utils.ToInt32(pStr)
		if err != nil {
			h.errorResponse(w, http.StatusUnprocessableEntity, "Invalid parameter type")
			return
		}
		if p > 0 {
			pageVal = p
		}
	}
	offsetVal := (pageVal - 1) * limitVal

	// Extract filter values
	genders, ok := filters["gender"].([]string)
	if !ok || genders == nil {
		genders = []string{}
	}
	ageGroup, _ := filters["age_group"].(string)
	minAge, _ := filters["min_age"].(int)
	maxAge, _ := filters["max_age"].(int)
	countryID, _ := filters["country_id"].(string)
	countryName, _ := filters["country_name"].(string)
	exactAge, _ := filters["age"].(int)

	countParams := db.CountProfilesAdvancedParams{
		Genders:     genders,
		AgeGroup:    ageGroup,
		CountryID:   countryID,
		CountryName: countryName,
		MinAge:      int32(minAge),
		MaxAge:      int32(maxAge),
		ExactAge:    int32(exactAge),
	}

	total, err := h.queries.CountProfilesAdvanced(r.Context(), countParams)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to count profiles")
		return
	}

	listParams := db.ListProfilesAdvancedParams{
		Genders:       genders,
		AgeGroup:      ageGroup,
		CountryID:     countryID,
		CountryName:   countryName,
		MinAge:        int32(minAge),
		MaxAge:        int32(maxAge),
		ExactAge:      int32(exactAge),
		SortBy:        "created_at",
		SortDirection: "desc",
		LimitVal:      limitVal,
		OffsetVal:     offsetVal,
	}

	profiles, err := h.queries.ListProfilesAdvanced(r.Context(), listParams)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to fetch profiles")
		return
	}

	data := []map[string]any{}
	for _, p := range profiles {
		data = append(data, map[string]any{
			"id":                  uuid.UUID(p.ID.Bytes).String(),
			"name":                p.Name,
			"gender":              p.Gender.String,
			"gender_probability":  p.GenderProbability.Float64,
			"age":                 p.Age.Int32,
			"age_group":           p.AgeGroup.String,
			"country_id":          p.CountryID.String,
			"country_name":        p.CountryName.String,
			"country_probability": p.CountryProbability.Float64,
			"created_at":          p.CreatedAt.Time.UTC().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"status": "success",
		"page":   pageVal,
		"limit":  limitVal,
		"total":  total,
		"data":   data,
	})
}

func (h *Handler) DeleteProfileByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "Missing profile ID")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid UUID format")
		return
	}

	// 1. Check if the profile exists
	_, err = h.queries.GetProfile(r.Context(), utils.ToUUID(id))
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "Profile not found")
		return
	}

	// 2. Perform the deletion
	err = h.queries.DeleteProfile(r.Context(), utils.ToUUID(id))
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Failed to delete profile")
		return
	}

	// 3. Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
