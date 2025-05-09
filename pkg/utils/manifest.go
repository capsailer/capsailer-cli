package utils

// Manifest represents the YAML configuration for Capsailer
type Manifest struct {
	Images []string `yaml:"images"`
	Charts []Chart  `yaml:"charts"`
}

// Chart represents a Helm chart and its configuration
type Chart struct {
	Name       string `yaml:"name"`
	Repo       string `yaml:"repo"`
	Version    string `yaml:"version"`
	ValuesFile string `yaml:"valuesFile,omitempty"`
}

// NewManifest creates a new empty manifest
func NewManifest() *Manifest {
	return &Manifest{
		Images: []string{},
		Charts: []Chart{},
	}
} 