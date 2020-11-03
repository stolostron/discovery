package ocm

// StandardKind ...
type StandardKind struct {
	Kind string `yaml:"kind,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Href string `href:"kind,omitempty"`
}

// Console ...
type Console struct {
	URL string `yaml:"url,omitempty"`
}
