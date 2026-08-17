package ui

// TextOpts Asking for text options
type TextOpts struct {
	Label   string
	Default string
	// ValidateFunc: method to validate input/default value
	ValidateFunc func(string) (bool, string, error)
}

// ChoiceOpts asking for choice options
type ChoiceOpts struct {
	Label   string
	Default int
	Choices []string
}
