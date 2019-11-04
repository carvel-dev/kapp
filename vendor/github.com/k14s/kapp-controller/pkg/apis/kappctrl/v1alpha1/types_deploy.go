package v1alpha1

type AppDeploy struct {
	Kapp *AppDeployKapp `json:"kapp,omitempty"`
}

type AppDeployKapp struct {
	IntoNs     string   `json:"intoNs,omitempty"`
	MapNs      []string `json:"mapNs,omitempty"`
	RawOptions []string `json:"rawOptions,omitempty"`
}
