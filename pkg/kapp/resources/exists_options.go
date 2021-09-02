package resources

type ExistsOpts struct {
	SameUID bool
}

func (e *ExistsOpts) checkForSameUID() bool {
	return e.SameUID
}

