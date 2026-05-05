package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// GenerateQueryKey takes any struct (like db.ListProfilesAdvancedParams)
// and generates a deterministic string key for Redis caching.
// This handles Query Normalization (Task 2) perfectly because the parsed filters
// are already canonicalized by the time they reach this point!
func GenerateQueryKey(prefix string, params interface{}) (string, error) {
	bytes, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	
	// Create an MD5 hash of the JSON to keep keys short and consistent
	hash := md5.Sum(bytes)
	return fmt.Sprintf("%s:%x", prefix, hash), nil
}
