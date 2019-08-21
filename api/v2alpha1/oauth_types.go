package v2alpha1

type OauthModeConfig struct {
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:UniqueItems=true
	Paths []Option `json:",inline"`
}

type Option struct {
	Path    *string  `json:"path"`
	Scopes  []string `json:"scopes,omitempty"`
	Methods []string `json:"methods,omitempty"`
}
