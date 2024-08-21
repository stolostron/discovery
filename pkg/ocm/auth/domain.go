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
	AuthMethod  string `json:"auth_method,omitempty" yaml:"auth_method,omitempty"`
	BaseURL     string `json:"base_url,omitempty" yaml:"base_url,omitempty"`
	BaseAuthURL string `json:"base_auth_url,omitempty" yaml:"base_auth_url,omitempty"`
	ID          string `json:"client_id,omitempty" yaml:"client_id,omitempty"`
	Secret      string `json:"client_secret,omitempty" yaml:"client_secret,omitempty"`
	Token       string `json:"ocmAPIToken,omitempty" yaml:"ocmAPIToken,omitempty"`
}

type AuthError struct {
	Code         int    `json:"code"`
	ErrorMessage string `json:"error"`
	Description  string `json:"error_description"`
	// Error is for setting an internal error for tracking
	Error error `json:"-"`
	// Response is for storing the raw response on an error
	Response []byte `json:"-"`
}
