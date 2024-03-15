package utils

// APISettings represents settings related to an API.
type APISettings struct {
	URL       string `json:"url,omitempty"`
	Listening string `json:"listening,omitempty"`
}

// Console represents settings related to a console.
type Console struct {
	URL string `yaml:"url,omitempty"`
}

// Metrics represents metrics related to a system or application.
type Metrics struct {
	OpenShiftVersion string `json:"openshift_version,omitempty"`
}

// StandardKind represents a standard kind with optional ID and Href.
type StandardKind struct {
	Kind string `yaml:"kind,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Href string `href:"kind,omitempty"`
}
