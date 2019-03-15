package diff

import (
	"fmt"
	"strings"

	"github.com/aryann/difflib"
	"github.com/fatih/color"
)

type TextDiffViewOpts struct {
	Context int // number of lines to show around changed lines; <0 for all
}

type TextDiffView struct {
	diff TextDiff
	opts TextDiffViewOpts
}

func NewTextDiffView(diff TextDiff, opts TextDiffViewOpts) TextDiffView {
	return TextDiffView{diff, opts}
}

func (v TextDiffView) String() string {
	lines := []string{}
	changedLines := map[int]struct{}{}

	for lineNum, diff := range v.diff {
		if diff.Delta != difflib.Common {
			changedLines[lineNum] = struct{}{}
		}
	}

	prevInContext := false
	emptyLineStr := "   "
	lineStr := func(line int) string { return fmt.Sprintf("%3d", line) }

	for lineNum, diff := range v.diff {
		switch diff.Delta {
		case difflib.RightOnly:
			lines = append(lines, color.New(color.FgGreen).Sprintf("%s %s + %s",
				emptyLineStr,
				lineStr(diff.LineRight),
				diff.Payload))

		case difflib.LeftOnly:
			lines = append(lines, color.New(color.FgRed).Sprintf("%s %s - %s",
				lineStr(diff.LineLeft),
				emptyLineStr,
				diff.Payload))

		case difflib.Common:
			newInContext := v.inContext(lineNum, changedLines)
			if lineNum != 0 && !prevInContext && newInContext {
				lines = append(lines, "  ...")
			}
			if newInContext {
				lines = append(lines, fmt.Sprintf("%s,%s   %s",
					lineStr(diff.LineLeft),
					lineStr(diff.LineRight),
					diff.Payload)) // LineLeft == LineRight
			}
			prevInContext = newInContext
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

func (v TextDiffView) inContext(lineNum int, changedLines map[int]struct{}) bool {
	if v.opts.Context < 0 {
		return true
	}
	for i := lineNum - v.opts.Context; i < lineNum+v.opts.Context; i++ {
		if _, found := changedLines[i]; found {
			return true
		}
	}
	return false
}
