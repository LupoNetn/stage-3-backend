package handlers

type GenderizeResp struct {
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
	Count       int32   `json:"count"`
}

type AgifyResp struct {
	Name string `json:"name"`
	Age  int32  `json:"age"`
}

type NationalizeResp struct {
	Name    string `json:"name"`
	Country []struct {
		CountryID   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}
