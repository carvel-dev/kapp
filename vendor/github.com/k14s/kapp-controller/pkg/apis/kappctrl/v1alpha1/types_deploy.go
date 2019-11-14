package v1alpha1

type AppDeploy struct {
	Kapp *AppDeployKapp `json:"kapp,omitempty"`
}

type AppDeployKapp struct {
	IntoNs     string   `json:"intoNs,omitempty"`
	MapNs      []string `json:"mapNs,omitempty"`
	RawOptions []string `json:"rawOptions,omitempty"`

	Inspect *AppDeployKappInspect `json:"inspect,omitempty"`
	Delete  *AppDeployKappDelete  `json:"delete,omitempty"`
}

type AppDeployKappInspect struct {
	RawOptions []string `json:"rawOptions,omitempty"`
}

type AppDeployKappDelete struct {
	RawOptions []string `json:"rawOptions,omitempty"`
}
