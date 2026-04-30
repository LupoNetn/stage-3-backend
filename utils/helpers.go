package utils

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToUUID converts a google/uuid to pgtype.UUID
func ToUUID(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}

// ToText converts a string to pgtype.Text
func ToText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// ToFloat8 converts a float64 to pgtype.Float8
func ToFloat8(f float64) pgtype.Float8 {
	return pgtype.Float8{Float64: f, Valid: true}
}

// ToInt4 converts an int32 to pgtype.Int4
func ToInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

// ToTimestamptz converts a time.Time to pgtype.Timestamptz
func ToTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// ToInt32 converts a string to int32
func ToInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	return int32(v), err
}

// ToFloat64 converts a string to float64
func ToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// ClassifyAgeGroup returns the age group category based on age
func ClassifyAgeGroup(age int32) string {
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

// NormalizeAndTokenize cleans and splits a string into tokens
func NormalizeAndTokenize(q string) []string {
	q = strings.ToLower(q)
	reg := regexp.MustCompile(`[^\w\s]`)
	q = reg.ReplaceAllString(q, "")
	q = strings.TrimSpace(q)
	return strings.Fields(q)
}

// AppendUnique appends a string to a slice if it's not already present
func AppendUnique(arr []string, val string) []string {
	for _, v := range arr {
		if v == val {
			return arr
		}
	}
	return append(arr, val)
}

// ClearCookie removes a cookie by setting it to the past
func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
