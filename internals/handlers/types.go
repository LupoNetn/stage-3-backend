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

// auth types
type GithubCLIAuth struct {
	Code         string `json:"code"`
	State        string `json:"state"`
	CodeVerifier string `json:"code_verifier"`
}

type GithubCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type GithubAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
}
