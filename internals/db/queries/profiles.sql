-- name: CreateProfile :one
INSERT INTO profiles (
    id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpsertUser :exec
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
    country_probability = EXCLUDED.country_probability;

-- name: GetProfile :one
SELECT * FROM profiles
WHERE id = $1 LIMIT 1;

-- name: GetProfileByName :one
SELECT * FROM profiles
WHERE name = $1 LIMIT 1;

-- name: ListProfilesAdvanced :many
SELECT * FROM profiles
WHERE
    (CARDINALITY(@genders::text[]) = 0 OR gender = ANY(@genders))
AND (@age_group::text = '' OR age_group = @age_group)
AND (@country_id::text = '' OR country_id ILIKE @country_id)
AND (@country_name::text = '' OR country_name ILIKE @country_name)
AND (@min_age::int = 0 OR age >= @min_age)
AND (@max_age::int = 0 OR age <= @max_age)
AND (@exact_age::int = 0 OR age = @exact_age)
AND (@min_gender_prob::float = 0 OR gender_probability >= @min_gender_prob)
AND (@min_country_prob::float = 0 OR country_probability >= @min_country_prob)
ORDER BY 
    CASE WHEN @sort_by::text = 'age' AND @sort_direction::text = 'asc' THEN age END ASC,
    CASE WHEN @sort_by::text = 'age' AND @sort_direction::text = 'desc' THEN age END DESC,
    CASE WHEN @sort_by::text = 'created_at' AND @sort_direction::text = 'asc' THEN created_at END ASC,
    CASE WHEN @sort_by::text = 'created_at' AND @sort_direction::text = 'desc' THEN created_at END DESC,
    CASE WHEN @sort_by::text = 'gender_probability' AND @sort_direction::text = 'asc' THEN gender_probability END ASC,
    CASE WHEN @sort_by::text = 'gender_probability' AND @sort_direction::text = 'desc' THEN gender_probability END DESC,
    created_at DESC
LIMIT @limit_val
OFFSET @offset_val;

-- name: CountProfilesAdvanced :one
SELECT COUNT(*) FROM profiles
WHERE
    (CARDINALITY(@genders::text[]) = 0 OR gender = ANY(@genders))
AND (@age_group::text = '' OR age_group = @age_group)
AND (@country_id::text = '' OR country_id ILIKE @country_id)
AND (@country_name::text = '' OR country_name ILIKE @country_name)
AND (@min_age::int = 0 OR age >= @min_age)
AND (@max_age::int = 0 OR age <= @max_age)
AND (@exact_age::int = 0 OR age = @exact_age)
AND (@min_gender_prob::float = 0 OR gender_probability >= @min_gender_prob)
AND (@min_country_prob::float = 0 OR country_probability >= @min_country_prob);

-- name: DeleteProfile :exec
DELETE FROM profiles
WHERE id = $1;
