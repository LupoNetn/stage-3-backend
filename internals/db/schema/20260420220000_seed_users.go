package migrations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSeedUsers, downSeedUsers)
}

func upSeedUsers(ctx context.Context, tx *sql.Tx) error {
	data, err := os.ReadFile("../../seed_profiles.json")
	if err != nil {
		
		data, err = os.ReadFile("./seed_profiles.json")
		if err != nil {
			fmt.Println("could not read json file for seeding db")
			return err
		}
	}

	type ProfileJSON struct {
		Name               string  `json:"name"`
		Gender             string  `json:"gender"`
		GenderProbability  float64 `json:"gender_probability"`
		Age                int32   `json:"age"`
		AgeGroup           string  `json:"age_group"`
		CountryID          string  `json:"country_id"`
		CountryName        string  `json:"country_name"`
		CountryProbability float64 `json:"country_probability"`
	}

	var root struct {
		Profiles []ProfileJSON `json:"profiles"`
	}

	if err := json.Unmarshal(data, &root); err != nil {
		fmt.Println("could not unmarshal json data to profile slice")
		return err
	}

	for _, p := range root.Profiles {
		uid, _ := uuid.NewV7()
		
		gender := sql.NullString{String: p.Gender, Valid: p.Gender != ""}
		genderProb := sql.NullFloat64{Float64: p.GenderProbability, Valid: true}
		age := sql.NullInt32{Int32: p.Age, Valid: true}
		ageGroup := sql.NullString{String: p.AgeGroup, Valid: p.AgeGroup != ""}
		countryID := sql.NullString{String: p.CountryID, Valid: p.CountryID != ""}
		countryName := sql.NullString{String: p.CountryName, Valid: p.CountryName != ""}
		countryProb := sql.NullFloat64{Float64: p.CountryProbability, Valid: true}

		_, err := tx.ExecContext(ctx, `
INSERT INTO profiles (
    id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (name) DO UPDATE SET
    gender = EXCLUDED.gender,
    gender_probability = EXCLUDED.gender_probability,
    age = EXCLUDED.age,
    age_group = EXCLUDED.age_group,
    country_id = EXCLUDED.country_id,
    country_name = EXCLUDED.country_name,
    country_probability = EXCLUDED.country_probability`,
			uid,
			p.Name,
			gender,
			genderProb,
			age,
			ageGroup,
			countryID,
			countryName,
			countryProb,
		)
		if err != nil {
			fmt.Printf("could not upsert user %s: %v\n", p.Name, err)
			return err
		}
	}

	return nil
}

func downSeedUsers(ctx context.Context, tx *sql.Tx) error {
	return nil
}
