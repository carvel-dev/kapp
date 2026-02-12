// Copyright 2012 Aryan Naraghi (aryan.naraghi@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package difflib provides functionality for computing the difference
// between two sequences of strings.
package difflib

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
)

// DeltaType describes the relationship of elements in two
// sequences. The following table provides a summary:
//
//	 Constant    Code   Meaning
//	----------  ------ ---------------------------------------
//	 Common      " "    The element occurs in both sequences.
//	 LeftOnly    "-"    The element is unique to sequence 1.
//	 RightOnly   "+"    The element is unique to sequence 2.
type DeltaType int

const (
	Common DeltaType = iota
	LeftOnly
	RightOnly
)

// String returns a string representation for DeltaType.
func (t DeltaType) String() string {
	switch t {
	case Common:
		return " "
	case LeftOnly:
		return "-"
	case RightOnly:
		return "+"
	}
	return "?"
}

type DiffRecord struct {
	Payload   string
	Delta     DeltaType
	LineLeft  int
	LineRight int
}

// String returns a string representation of d. The string is a
// concatenation of the delta type and the payload.
func (d DiffRecord) String() string {
	return fmt.Sprintf("%s %s", d.Delta, d.Payload)
}

// Diff returns the result of diffing the seq1 and seq2.
func Diff(seq1, seq2 []string) (diff []DiffRecord) {
	// Trims any common elements at the heads and tails of the
	// sequences before running the diff algorithm. This is an
	// optimization.
	start, end := numEqualStartAndEndElements(seq1, seq2)

	for i, content := range seq1[:start] {
		diff = append(diff, DiffRecord{content, Common, i, i})
	}

	diffRes := compute(seq1[start:len(seq1)-end], seq2[start:len(seq2)-end], start)
	diff = append(diff, diffRes...)

	for i, content := range seq1[len(seq1)-end:] {
		diff = append(diff, DiffRecord{content, Common, len(seq1) - end + i, len(seq2) - end + i})
	}
	return
}

// HTMLDiff returns the results of diffing seq1 and seq2 as an HTML
// string. The resulting HTML is a table without the opening and
// closing table tags. Each table row represents a DiffRecord. The
// first and last columns contain the "line numbers" for seq1 and
// seq2, respectively (the function assumes that seq1 and seq2
// represent the lines in a file). The second and third columns
// contain the actual file contents.
//
// The cells that contain line numbers are decorated with the class
// "line-num". The cells that contain deleted elements are decorated
// with "deleted" and the cells that contain added elements are
// decorated with "added".
func HTMLDiff(seq1, seq2 []string) string {
	buf := bytes.NewBufferString("")
	i, j := 0, 0
	for _, d := range Diff(seq1, seq2) {
		buf.WriteString(`<tr><td class="line-num">`)
		if d.Delta == Common || d.Delta == LeftOnly {
			i++
			fmt.Fprintf(buf, "%d</td><td", i)
			if d.Delta == LeftOnly {
				fmt.Fprint(buf, ` class="deleted"`)
			}
			fmt.Fprintf(buf, "><pre>%s</pre>", d.Payload)
		} else {
			buf.WriteString("</td><td>")
		}
		buf.WriteString("</td><td")
		if d.Delta == Common || d.Delta == RightOnly {
			j++
			if d.Delta == RightOnly {
				fmt.Fprint(buf, ` class="added"`)
			}
			fmt.Fprintf(buf, `><pre>%s</pre></td><td class="line-num">%d`, d.Payload, j)
		} else {
			buf.WriteString(`></td><td class="line-num">`)
		}
		buf.WriteString("</td></tr>\n")
	}
	return buf.String()
}

// PPDiff returns the results of diffing left and right as an pretty
// printed string. It will display all the lines of both the sequences
// that are being compared.
// When the left is different from right it will prepend a " - |" before
// the line.
// When the right is different from left it will prepend a " + |" before
// the line.
// When the right and left are equal it will prepend a "   |" before
// the line.
func PPDiff(left, right []string) string {
	var sb strings.Builder

	recs := Diff(right, left)

	for _, diff := range recs {
		var mark string

		switch diff.Delta {
		case RightOnly:
			mark = " + |"
		case LeftOnly:
			mark = " - |"
		case Common:
			mark = "   |"
		}

		// make sure to have line numbers to make sure diff is truly unique
		sb.WriteString(fmt.Sprintf("%3d,%3d%s%s\n", diff.LineLeft, diff.LineRight, mark, diff.Payload))
	}

	return sb.String()
}

// numEqualStartAndEndElements returns the number of elements a and b
// have in common from the beginning and from the end. If a and b are
// equal, start will equal len(a) == len(b) and end will be zero.
func numEqualStartAndEndElements(seq1, seq2 []string) (start, end int) {
	for start < len(seq1) && start < len(seq2) && seq1[start] == seq2[start] {
		start++
	}
	i, j := len(seq1)-1, len(seq2)-1
	for i > start && j > start && seq1[i] == seq2[j] {
		i--
		j--
		end++
	}
	return
}

// intMatrix returns a 2-dimensional slice of ints with the given
// number of rows and columns.
func intMatrix(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]int, cols)
	}
	return matrix
}

// longestCommonSubsequenceMatrix returns the table that results from
// applying the dynamic programming approach for finding the longest
// common subsequence of seq1 and seq2.
func longestCommonSubsequenceMatrix(seq1, seq2 []string) [][]int {
	matrix := intMatrix(len(seq1)+1, len(seq2)+1)
	for i := 1; i < len(matrix); i++ {
		for j := 1; j < len(matrix[i]); j++ {
			if seq1[len(seq1)-i] == seq2[len(seq2)-j] {
				matrix[i][j] = matrix[i-1][j-1] + 1
			} else {
				matrix[i][j] = int(math.Max(float64(matrix[i-1][j]),
					float64(matrix[i][j-1])))
			}
		}
	}
	return matrix
}

// compute is the unexported helper for Diff that returns the results of
// diffing left and right.
func compute(seq1, seq2 []string, startLine int) (diff []DiffRecord) {
	matrix := longestCommonSubsequenceMatrix(seq1, seq2)
	i, j := len(seq1), len(seq2)
	for i > 0 || j > 0 {
		if i > 0 && matrix[i][j] == matrix[i-1][j] {
			diff = append(diff, DiffRecord{seq1[len(seq1)-i], LeftOnly, startLine + len(seq1) - i, startLine + len(seq2) - j})
			i--
		} else if j > 0 && matrix[i][j] == matrix[i][j-1] {
			diff = append(diff, DiffRecord{seq2[len(seq2)-j], RightOnly, startLine + len(seq1) - i, startLine + len(seq2) - j})
			j--
		} else if i > 0 && j > 0 {
			diff = append(diff, DiffRecord{seq1[len(seq1)-i], Common, startLine + len(seq1) - i, startLine + len(seq2) - j})
			i--
			j--
		}
	}
	return
}

// A pair is a pair of values tracked for both the x and y side of a diff.
// It is typically a pair of line indexes.
type pair struct{ x, y int }

// Diff returns an anchored diff of the two texts old and new
// in the “unified diff” format. If old and new are identical,
// Diff returns a nil slice (no output).
//
// Unix diff implementations typically look for a diff with
// the smallest number of lines inserted and removed,
// which can in the worst case take time quadratic in the
// number of lines in the texts. As a result, many implementations
// either can be made to run for a long time or cut off the search
// after a predetermined amount of work.
//
// In contrast, this implementation looks for a diff with the
// smallest number of “unique” lines inserted and removed,
// where unique means a line that appears just once in both old and new.
// We call this an “anchored diff” because the unique lines anchor
// the chosen matching regions. An anchored diff is usually clearer
// than a standard diff, because the algorithm does not try to
// reuse unrelated blank lines or closing braces.
// The algorithm also guarantees to run in O(n log n) time
// instead of the standard O(n²) time.
//
// Some systems call this approach a “patience diff,” named for
// the “patience sorting” algorithm, itself named for a solitaire card game.
// We avoid that name for two reasons. First, the name has been used
// for a few different variants of the algorithm, so it is imprecise.
// Second, the name is frequently interpreted as meaning that you have
// to wait longer (to be patient) for the diff, meaning that it is a slower algorithm,
// when in fact the algorithm is faster than the standard one.
func AnchoredDiff(seq1, seq2 []string) []DiffRecord {
	diff := []DiffRecord{}
	equalDiff := []DiffRecord{}

	// Loop over matches to consider,
	// expanding each match to include surrounding lines,
	// and then printing diff chunks.
	// To avoid setup/teardown cases outside the loop,
	// tgs returns a leading {0,0} and trailing {len(x), len(y)} pair
	// in the sequence of matches.
	var (
		done  pair       // printed up to x[:done.x] and y[:done.y]
		chunk pair       // start lines of current chunk
		count pair       // number of lines from each side in current chunk
		ctext []struct{} // lines for current chunk
	)
	for _, m := range tgs(seq1, seq2) {
		if m.x < done.x {
			// Already handled scanning forward from earlier match.
			continue
		}

		// Expand matching lines as far possible,
		// establishing that x[start.x:end.x] == y[start.y:end.y].
		// Note that on the first (or last) iteration we may (or definitely do)
		// have an empty match: start.x==end.x and start.y==end.y.
		start := m
		for start.x > done.x && start.y > done.y && seq1[start.x-1] == seq2[start.y-1] {
			start.x--
			start.y--
		}
		end := m
		for end.x < len(seq1) && end.y < len(seq2) && seq1[end.x] == seq2[end.y] {
			equalDiff = append(equalDiff, DiffRecord{seq1[end.x], Common, end.x, end.y})
			end.x++
			end.y++
		}

		// If both sequences are identical, then add 'common' diff for all lines
		if start.x == 0 && start.y == 0 && end.x == len(seq1) && end.y == len(seq2) {
			diff = append(diff, equalDiff...)
		}

		// Emit the mismatched lines before start into this chunk.
		// (No effect on first sentinel iteration, when start = {0,0}.)
		for _, s := range seq1[done.x:start.x] {
			diff = append(diff, DiffRecord{s, LeftOnly, chunk.x + count.x, chunk.y + count.y})
			ctext = append(ctext, struct{}{})
			count.x++
		}
		for _, s := range seq2[done.y:start.y] {
			diff = append(diff, DiffRecord{s, RightOnly, chunk.x + count.x, chunk.y + count.y})
			ctext = append(ctext, struct{}{})
			count.y++
		}

		// If we're not at EOF and have too few common lines,
		// the chunk includes all the common lines and continues.
		const C = 30 // maximum number of context lines
		if (end.x < len(seq1) || end.y < len(seq2)) &&
			(end.x-start.x < C || (len(ctext) > 0 && end.x-start.x < 2*C)) {
			for _, s := range seq1[start.x:end.x] {
				ctext = append(ctext, struct{}{})
				diff = append(diff, DiffRecord{s, Common, chunk.x + count.x, chunk.y + count.y})
				count.x++
				count.y++
			}
			done = end
			continue
		}

		// End chunk with common lines for context.
		if len(ctext) > 0 {
			n := end.x - start.x
			if n > C {
				n = C
			}
			for _, s := range seq1[start.x : start.x+n] {
				ctext = append(ctext, struct{}{})
				diff = append(diff, DiffRecord{s, Common, chunk.x + count.x, chunk.y + count.y})
				count.x++
				count.y++
			}
			done = pair{start.x + n, start.y + n}

			// Format and emit chunk.
			// Convert line numbers to 1-indexed.
			// Special case: empty file shows up as 0,0 not 1,0.
			if count.x > 0 {
				chunk.x++
			}
			if count.y > 0 {
				chunk.y++
			}
			count.x = 0
			count.y = 0
			ctext = ctext[:0]
		}

		// If we reached EOF, we're done.
		if end.x >= len(seq1) && end.y >= len(seq2) {
			break
		}

		// Otherwise start a new chunk.
		chunk = pair{end.x - C, end.y - C}
		for _, s := range seq1[chunk.x:end.x] {
			ctext = append(ctext, struct{}{})
			diff = append(diff, DiffRecord{s, Common, chunk.x + count.x, chunk.y + count.y})
			count.x++
			count.y++
		}
		done = end
	}

	return diff
}

// tgs returns the pairs of indexes of the longest common subsequence
// of unique lines in x and y, where a unique line is one that appears
// once in x and once in y.
//
// The longest common subsequence algorithm is as described in
// Thomas G. Szymanski, “A Special Case of the Maximal Common
// Subsequence Problem,” Princeton TR #170 (January 1975),
// available at https://research.swtch.com/tgs170.pdf.
func tgs(x, y []string) []pair {
	// Count the number of times each string appears in a and b.
	// We only care about 0, 1, many, counted as 0, -1, -2
	// for the x side and 0, -4, -8 for the y side.
	// Using negative numbers now lets us distinguish positive line numbers later.
	m := make(map[string]int)
	for _, s := range x {
		if c := m[s]; c > -2 {
			m[s] = c - 1
		}
	}
	for _, s := range y {
		if c := m[s]; c > -8 {
			m[s] = c - 4
		}
	}

	// Now unique strings can be identified by m[s] = -1+-4.
	//
	// Gather the indexes of those strings in x and y, building:
	//	xi[i] = increasing indexes of unique strings in x.
	//	yi[i] = increasing indexes of unique strings in y.
	//	inv[i] = index j such that x[xi[i]] = y[yi[j]].
	var xi, yi, inv []int
	for i, s := range y {
		if m[s] == -1+-4 {
			m[s] = len(yi)
			yi = append(yi, i)
		}
	}
	for i, s := range x {
		if j, ok := m[s]; ok && j >= 0 {
			xi = append(xi, i)
			inv = append(inv, j)
		}
	}

	// Apply Algorithm A from Szymanski's paper.
	// In those terms, A = J = inv and B = [0, n).
	// We add sentinel pairs {0,0}, and {len(x),len(y)}
	// to the returned sequence, to help the processing loop.
	J := inv
	n := len(xi)
	T := make([]int, n)
	L := make([]int, n)
	for i := range T {
		T[i] = n + 1
	}
	for i := 0; i < n; i++ {
		k := sort.Search(n, func(k int) bool {
			return T[k] >= J[i]
		})
		T[k] = J[i]
		L[i] = k + 1
	}
	k := 0
	for _, v := range L {
		if k < v {
			k = v
		}
	}
	seq := make([]pair, 2+k)
	seq[1+k] = pair{len(x), len(y)} // sentinel at end
	lastj := n
	for i := n - 1; i >= 0; i-- {
		if L[i] == k && J[i] < lastj {
			seq[k] = pair{xi[i], yi[J[i]]}
			k--
		}
	}
	seq[0] = pair{0, 0} // sentinel at start
	return seq
}
