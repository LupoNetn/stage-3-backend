package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/luponetn/hng-stage-1/internals/config"
	"github.com/luponetn/hng-stage-1/internals/db"
	"github.com/google/uuid"
)

type SeedData struct {
	Profiles []struct {
		Name               string  `json:"name"`
		Gender             string  `json:"gender"`
		GenderProbability  float64 `json:"gender_probability"`
		Age                int32   `json:"age"`
		AgeGroup           string  `json:"age_group"`
		CountryID          string  `json:"country_id"`
		CountryName        string  `json:"country_name"`
		CountryProbability float64 `json:"country_probability"`
	} `json:"profiles"`
}

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()

	pool, err := db.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)

	data, err := os.ReadFile("seed_profiles.json")
	if err != nil {
		log.Fatalf("failed to read seed file: %v", err)
	}

	var seed SeedData
	if err := json.Unmarshal(data, &seed); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	fmt.Printf("Seeding %d profiles...\n", len(seed.Profiles))

	ctx := context.Background()
	count := 0
	for _, p := range seed.Profiles {
		id, err := uuid.NewV7()
		if err != nil {
			continue
		}

		arg := db.UpsertUserParams{
			ID: pgtype.UUID{
				Bytes: id,
				Valid: true,
			},
			Name: p.Name,
			Gender: pgtype.Text{
				String: p.Gender,
				Valid:  true,
			},
			GenderProbability: pgtype.Float8{
				Float64: p.GenderProbability,
				Valid:   true,
			},
			Age: pgtype.Int4{
				Int32: p.Age,
				Valid: true,
			},
			AgeGroup: pgtype.Text{
				String: p.AgeGroup,
				Valid:  true,
			},
			CountryID: pgtype.Text{
				String: p.CountryID,
				Valid:  true,
			},
			CountryName: pgtype.Text{
				String: p.CountryName,
				Valid:  true,
			},
			CountryProbability: pgtype.Float8{
				Float64: p.CountryProbability,
				Valid:   true,
			},
		}

		err = queries.UpsertUser(ctx, arg)
		if err != nil {
			continue
		}
		count++
	}

	fmt.Printf("Successfully seeded %d profiles.\n", count)
}
