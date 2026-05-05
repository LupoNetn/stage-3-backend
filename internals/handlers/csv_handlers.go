package handlers

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luponetn/hng-stage-1/utils"
)

type ImportSummary struct {
	Status    string `json:"status"`
	TotalRows int    `json:"total_rows"`
	Inserted  int    `json:"inserted"`
	Skipped   int    `json:"skipped"`
	Reasons   struct {
		DuplicateName int `json:"duplicate_name"`
		InvalidAge    int `json:"invalid_age"`
		MissingFields int `json:"missing_fields"`
		MalformedRow  int `json:"malformed_row"`
	} `json:"reasons"`
}

func (h *Handler) ImportProfilesCSV(w http.ResponseWriter, r *http.Request) {
	// 1. Setup the response struct
	summary := ImportSummary{
		Status: "success",
	}

	// 2. We use multipart reader for streaming upload
	err := r.ParseMultipartForm(10 << 20) // 10MB limit in memory, rest to disk
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// 3. Initialize CSV Reader
	reader := csv.NewReader(file)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Failed to read CSV header")
		return
	}

	// Map headers to indices
	headerMap := make(map[string]int)
	for i, col := range header {
		headerMap[strings.ToLower(strings.TrimSpace(col))] = i
	}

	// Ensure required columns exist
	nameIdx, nameOk := headerMap["name"]
	if !nameOk {
		h.errorResponse(w, http.StatusBadRequest, "Missing 'name' column")
		return
	}

	genderIdx, genderOk := headerMap["gender"]
	ageIdx, ageOk := headerMap["age"]
	countryIdx, countryOk := headerMap["country_id"]

	// 4. Batching setup
	const batchSize = 1000
	var batch []map[string]any

	// Helper to flush batch to DB
	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		b := &pgx.Batch{}
		for _, row := range batch {
			id, _ := uuid.NewV7()
			b.Queue(
				`INSERT INTO profiles (id, name, gender, age, age_group, country_id, country_name) 
				 VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (name) DO NOTHING`,
				id, row["name"], row["gender"], row["age"], row["age_group"], row["country_id"], row["country_name"],
			)
		}

		br := h.pool.SendBatch(r.Context(), b)
		defer br.Close()

		insertedThisBatch := 0
		for i := 0; i < len(batch); i++ {
			ct, err := br.Exec()
			if err != nil {
				// We shouldn't fail the whole upload, but tracking individual insert failures in a batch
				// if DO NOTHING doesn't trigger, means it's malformed DB side.
				summary.Skipped++
				summary.Reasons.MalformedRow++
				continue
			}
			
			if ct.RowsAffected() == 0 {
				summary.Skipped++
				summary.Reasons.DuplicateName++
			} else {
				insertedThisBatch++
			}
		}
		summary.Inserted += insertedThisBatch
		batch = nil // Reset batch
	}

	// 5. Stream processing
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		summary.TotalRows++

		if err != nil {
			summary.Skipped++
			summary.Reasons.MalformedRow++
			continue
		}

		// Validation
		name := strings.TrimSpace(record[nameIdx])
		if name == "" {
			summary.Skipped++
			summary.Reasons.MissingFields++
			continue
		}

		var age int
		if ageOk && len(record) > ageIdx {
			age, err = strconv.Atoi(strings.TrimSpace(record[ageIdx]))
			if err != nil || age < 0 {
				summary.Skipped++
				summary.Reasons.InvalidAge++
				continue
			}
		}

		gender := ""
		if genderOk && len(record) > genderIdx {
			gender = strings.ToLower(strings.TrimSpace(record[genderIdx]))
			if gender != "male" && gender != "female" && gender != "" {
				summary.Skipped++
				summary.Reasons.MalformedRow++
				continue
			}
		}

		countryID := ""
		if countryOk && len(record) > countryIdx {
			countryID = strings.ToUpper(strings.TrimSpace(record[countryIdx]))
		}

		ageGroup := utils.ClassifyAgeGroup(int32(age))
		countryName := utils.GetCountryName(countryID)

		// Add to batch
		batch = append(batch, map[string]any{
			"name":         name,
			"gender":       gender,
			"age":          age,
			"age_group":    ageGroup,
			"country_id":   countryID,
			"country_name": countryName,
		})

		// Flush if batch full
		if len(batch) >= batchSize {
			flushBatch()
		}
	}

	// Flush remaining
	flushBatch()

	// 6. Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}
