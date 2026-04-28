package utils

import (
	"strconv"
	"strings"
)

// CreateRuleBasedFilter interprets search tokens into a structured filter map.
func CreateRuleBasedFilter(tokens []string) (map[string]any, bool) {
	filters := make(map[string]any)

	type Filter struct {
		Genders   []string
		Age       int
		AgeGroup  string
		MinAge    int
		MaxAge    int
		CountryID string
		Country   string
	}

	f := Filter{
		Genders: []string{},
	}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

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

		if i+2 < len(tokens) && (token == "male" || token == "female") && tokens[i+1] == "and" {
			g1 := normalizeGender(tokens[i])
			g2 := normalizeGender(tokens[i+2])
			f.Genders = AppendUnique(f.Genders, g1)
			f.Genders = AppendUnique(f.Genders, g2)
			i += 2
			continue
		}

		g := normalizeGender(token)
		if g != "" {
			f.Genders = AppendUnique(f.Genders, g)
		}

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
		if f.AgeGroup != "" {
			filters["age_group"] = f.AgeGroup
		}
		filters["min_age"] = 0
		filters["max_age"] = 0
	}

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
