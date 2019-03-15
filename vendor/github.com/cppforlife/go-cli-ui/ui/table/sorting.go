package table

import (
	"math"
)

type Sorting struct {
	SortBy []ColumnSort
	Rows   [][]Value
}

func (s Sorting) Len() int { return len(s.Rows) }

func (s Sorting) Less(i, j int) bool {
	var leftScore, rightScore float64

	for ci, cs := range s.SortBy {
		var left, right Value

		left = s.Rows[i][cs.Column].Value()
		right = s.Rows[j][cs.Column].Value()

		c := left.Compare(right)

		if c == 0 {
			leftScore += 1000 * math.Pow10(10-ci)
			rightScore += 1000 * math.Pow10(10-ci)
		} else {
			if (cs.Asc && c == -1) || (!cs.Asc && c == 1) {
				leftScore += 1000 * math.Pow10(10-ci)
			} else {
				rightScore += 1000 * math.Pow10(10-ci)
			}
		}
	}

	return leftScore > rightScore
}

func (s Sorting) Swap(i, j int) { s.Rows[i], s.Rows[j] = s.Rows[j], s.Rows[i] }
