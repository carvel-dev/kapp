package patch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch/patch"
)

var _ = Describe("ArrayIndex", func() {
	dummyPath := MustNewPointerFromString("")

	Describe("Concrete", func() {
		It("returns positive index", func() {
			idx := ArrayIndex{Index: 0, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 1, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 2, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))
		})

		It("wraps around negative index one time", func() {
			idx := ArrayIndex{Index: -0, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: -1, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))

			idx = ArrayIndex{Index: -2, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: -3, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))
		})

		It("does not work with empty arrays", func() {
			idx := ArrayIndex{Index: 0, Modifiers: nil, Array: []interface{}{}, Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{0, []interface{}{}, dummyPath}))

			p := PrevModifier{}
			n := NextModifier{}

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, n}, Array: []interface{}{}, Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{0, []interface{}{}, dummyPath}))
		})

		It("does not work with index out of bounds", func() {
			idx := ArrayIndex{Index: 3, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{3, []interface{}{1, 2, 3}, dummyPath}))

			idx = ArrayIndex{Index: -4, Modifiers: nil, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{-4, []interface{}{1, 2, 3}, dummyPath}))
		})

		It("returns previous item when previous modifier is used", func() {
			p := PrevModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p, p, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{-4, []interface{}{1, 2, 3}, dummyPath}))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{p, p, p, p, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{-5, []interface{}{1, 2, 3}, dummyPath}))

			idx = ArrayIndex{Index: 2, Modifiers: []Modifier{p, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))
		})

		It("returns next item when next modifier is used", func() {
			n := NextModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{n}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, n}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{3, []interface{}{1, 2, 3}, dummyPath}))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, n, n}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			_, err = idx.Concrete()
			Expect(err).To(Equal(OpMissingIndexErr{4, []interface{}{1, 2, 3}, dummyPath}))
		})

		It("works with multiple previous and next modifiers", func() {
			p := PrevModifier{}
			n := NextModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{p, n}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(0))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(1))

			idx = ArrayIndex{Index: 0, Modifiers: []Modifier{n, n, n, p}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			Expect(idx.Concrete()).To(Equal(2))
		})

		It("does not support any other modifier except previous and next", func() {
			b := BeforeModifier{}

			idx := ArrayIndex{Index: 0, Modifiers: []Modifier{b}, Array: []interface{}{1, 2, 3}, Path: dummyPath}
			_, err := idx.Concrete()
			Expect(err.Error()).To(Equal("Expected to find one of the following modifiers: 'prev', 'next', but found modifier 'patch.BeforeModifier'"))
		})
	})
})
