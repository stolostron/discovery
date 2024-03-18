package utils

// APISettings represents settings related to an API.
type APISettings struct {
	URL       string `json:"url,omitempty" yaml:"url,omitempty"`
	Listening string `json:"listening,omitempty" yaml:"listening,omitempty"`
}

// Console represents settings related to a console.
type Console struct {
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Metrics represents metrics related to a system or application.
type Metrics struct {
	OpenShiftVersion string `json:"openshift_version,omitempty" yaml:"openshift_version,omitempty"`
}

// StandardKind represents a standard kind with optional ID and Href.
type StandardKind struct {
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	ID   string `json:"id,omitempty" yaml:"id,omitempty"`
	Href string `json:"href,omitempty" yaml:"href,omitempty"`
}
