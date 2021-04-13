// Copyright Contributors to the Open Cluster Management project

package auth

// AuthTokenResponse ...
type AuthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	IDToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

type AuthRequest struct {
	BaseURL string
	Token   string
}

type AuthError struct {
	Code         int    `json:"code"`
	ErrorMessage string `json:"error"`
	Description  string `json:"error_description"`
	Error        error  `json:"-"`
	Response     []byte `json:"-"`
}
