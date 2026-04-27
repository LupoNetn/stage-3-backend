package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luponetn/hng-stage-1/internals/db"
)

// Helper methods for the Handler
func classifyAgeGroup(age int32) string {
	switch {
	case age <= 12:
		return "child"
	case age <= 19:
		return "teenager"
	case age <= 59:
		return "adult"
	default:
		return "senior"
	}
}

func mapToStruct(src any, dest any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func (h *Handler) profileResponse(w http.ResponseWriter, status int, msg string, p db.Profile) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

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

	resp := map[string]any{
		"status": "success",
		"data":   data,
	}
	if msg != "success" {
		resp["message"] = msg
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}

// pgtype conversion helpers
func toUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}
func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}
func toFloat8(f float64) pgtype.Float8 {
	return pgtype.Float8{Float64: f, Valid: true}
}
func toInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}
func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func toInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

func toFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func NormalizeAndTokenize(q string) []string {
	q = strings.ToLower(q)

	reg := regexp.MustCompile(`[^\w\s]`)
	q = reg.ReplaceAllString(q, "")

	q = strings.TrimSpace(q)

	tokens := strings.Fields(q)

	return tokens
}



func CreateRuleBasedFilter(tokens []string) (map[string]any, bool) {
	filters := make(map[string]any)

	type Filter struct {
		Genders     []string
		Age         int
		AgeGroup    string
		MinAge      int
		MaxAge      int
		CountryID   string
		Country string
	}

	f := Filter{
		Genders: []string{},
	}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		// location filter: "from <token>" or "in <token>"
		// token < 3 chars → treat as country code (e.g. "ng"), else → country name
		if (token == "from" || token == "in") && i+1 < len(tokens) {
			loc := strings.ToLower(tokens[i+1])
			if len(loc) < 3 {
				f.CountryID = loc
			} else {
				f.Country = loc
			}
			i++
			continue
		}

		//gender filter
		//"male and female" pattern matching
		if i+2 < len(tokens) && (token == "male" || token == "female") && tokens[i+1] == "and" {
			g1 := normalizeGender(tokens[i])
			g2 := normalizeGender(tokens[i+2])

			f.Genders = appendUnique(f.Genders, g1)
			f.Genders = appendUnique(f.Genders, g2)

			i += 2
			continue
		}

		g := normalizeGender(token)
		if g != "" {
			f.Genders = appendUnique(f.Genders, g)
		}

		//age and age group filter
		if i+3 < len(tokens) && (token == "not") && (tokens[i+1] == "older") && (tokens[i+2] == "than") {
			f.MaxAge, _ = strconv.Atoi(tokens[i+3])
			i += 3
			continue
		}

		if i+3 < len(tokens) && (token == "not") && (tokens[i+1] == "younger") && (tokens[i+2] == "than") {
			f.MinAge, _ = strconv.Atoi(tokens[i+3])
			i += 3
			continue
		}

		if i+2 < len(tokens) && (token == "older") && (tokens[i+1] == "than") {
			f.MinAge, _ = strconv.Atoi(tokens[i+2])
			i += 2
			continue
		}

		if i+2 < len(tokens) && (token == "younger") && (tokens[i+1] == "than") {
			f.MaxAge, _ = strconv.Atoi(tokens[i+2])
			i += 2
			continue
		}

		if i+1 < len(tokens) && (token == "above") {
			f.MinAge, _ = strconv.Atoi(tokens[i+1])
			i += 1
			continue
		}

		if i+1 < len(tokens) && (token == "below") {
			f.MaxAge, _ = strconv.Atoi(tokens[i+1])
			i += 1
			continue
		}

		// "young" → ages 16-24 (not a stored age group per spec)
		if token == "young" {
			if f.MinAge == 0 {
				f.MinAge = 16
			}
			if f.MaxAge == 0 {
				f.MaxAge = 24
			}
			continue
		}

		if token == "adult" {
			f.AgeGroup = "adult"
		}

		ageGroup := normalizeAgeGroup(token)
		if ageGroup != "" {
			f.AgeGroup = ageGroup
		}

		age, err := strconv.Atoi(token)
		if err == nil {
			f.Age = age
		}
	}

	filters["gender"] = f.Genders
	filters["age_group"] = f.AgeGroup
	filters["min_age"] = f.MinAge
	filters["max_age"] = f.MaxAge
	filters["country_id"] = f.CountryID
	filters["country_name"] = f.Country
	filters["age"] = f.Age

	if f.MinAge != 0 && f.MaxAge != 0 {
		filters["age_group"] = ""
	}

	if f.MaxAge != 0 && f.MinAge > f.MaxAge {
		//add age group back
		if f.AgeGroup != "" {
			filters["age_group"] = f.AgeGroup
		}
		filters["min_age"] = 0
		filters["max_age"] = 0
	}

	// interpretable if at least one filter is meaningful
	genders := filters["gender"].([]string)
	interpretable := len(genders) > 0 ||
		filters["age_group"].(string) != "" ||
		filters["min_age"].(int) != 0 ||
		filters["max_age"].(int) != 0 ||
		filters["country_id"].(string) != "" ||
		filters["country_name"].(string) != "" ||
		filters["age"].(int) != 0

	return filters, interpretable

}

func normalizeGender(t string) string {
	switch t {
	case "male", "males", "man", "men", "boy", "boys":
		return "male"
	case "female", "females", "woman", "women", "girl", "girls":
		return "female"
	}
	return ""
}

func normalizeAgeGroup(t string) string {
	switch t {
	case "child", "little", "baby":
		return "child"
	case "teenager", "teen", "teens", "youngins", "teenagers":
		return "teenager"
	case "adult":
		return "adult"
	case "elderly", "older", "old", "aged", "senior":
		return "senior"
	}
	return ""
}

func appendUnique(arr []string, val string) []string {
	for _, v := range arr {
		if v == val {
			return arr
		}
	}
	return append(arr, val)
}
