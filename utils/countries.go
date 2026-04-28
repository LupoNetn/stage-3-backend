package utils

// GetCountryName returns the full name of a country given its ISO 2-letter code.
func GetCountryName(code string) string {
	countries := map[string]string{
		"US": "United States",
		"GB": "United Kingdom",
		"CA": "Canada",
		"NG": "Nigeria",
		"GH": "Ghana",
		"KE": "Kenya",
		"ZA": "South Africa",
		"IN": "India",
		"DE": "Germany",
		"FR": "France",
		"CN": "China",
		"JP": "Japan",
		"BR": "Brazil",
		"AU": "Australia",
		"RU": "Russia",
		"ES": "Spain",
		"IT": "Italy",
		"NL": "Netherlands",
		"MX": "Mexico",
	}

	if name, ok := countries[code]; ok {
		return name
	}
	return code // Fallback to code if name is unknown
}
