package app

import "time"

type AppFilter struct {
	CreatedAtBeforeTime *time.Time
	CreatedAtAfterTime  *time.Time

	Labels []string
}
