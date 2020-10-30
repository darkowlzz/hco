package v1

// ImageReference contains image references to all the components.
type ImageReference struct {
	// +kubebuilder:validation:Optional
	App string `json:"app,omitempty"`

	// +kubebuilder:validation:Optional
	SidecarA string `json:"sidecarA,omitempty"`

	// +kubebuilder:validation:Optional
	SidecarB string `json:"sidecarB,omitempty"`
}
