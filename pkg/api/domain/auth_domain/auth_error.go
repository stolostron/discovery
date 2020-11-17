package auth_domain

type AuthError struct {
	Code         int    `json:"code"`
	ErrorMessage string `json:"error"`
	Description  string `json:"error_description"`
	Error        error  `json:"-"`
	Response     []byte `json:"-"`
	// reason       string
	// message      string
}

// func (e *AuthError) Reason() string {
// 	return e.reason
// }
// func (e *AuthError) Message() string {
// 	return e.message
// }
