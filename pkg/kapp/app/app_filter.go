package app

import "time"

type AppFilter struct {
	CreatedAtBeforeTime *time.Time
	CreatedAtAfterTime  *time.Time

	Labels []string
}

func (f AppFilter) Apply(apps []App) []App {
	var result []App

	for _, app := range apps {
		if f.Matches(app) {
			result = append(result, app)
		}
	}
	return result
}

func (f AppFilter) Matches(app App) bool {
	if f.CreatedAtBeforeTime != nil {
		if lc, _ := app.LastChange(); lc.Meta().StartedAt.After(*f.CreatedAtBeforeTime) {
			return false
		}
	}

	if f.CreatedAtAfterTime != nil {
		if lc, _ := app.LastChange(); lc.Meta().StartedAt.Before(*f.CreatedAtAfterTime) {
			return false
		}
	}
	return true
}
