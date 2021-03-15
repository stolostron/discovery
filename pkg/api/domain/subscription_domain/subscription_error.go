package subscription_domain

// SubscriptionError represents the error format response by OCM on a subscription request
type SubscriptionError struct {
	Kind     string `json:"kind"`
	ID       string `json:"id"`
	Href     string `json:"href"`
	Code     string `json:"code"`
	Reason   string `json:"reason"`
	Error    error  `json:"-"`
	Response []byte `json:"-"`
}
