package patch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch/patch"
)

var _ = Describe("FindOp.Apply", func() {
	It("returns document if path is for the entire document", func() {
		res, err := FindOp{Path: MustNewPointerFromString("")}.Apply("a")
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal("a"))
	})

	Describe("array item", func() {
		It("finds array item", func() {
			res, err := FindOp{Path: MustNewPointerFromString("/0")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(1))

			res, err = FindOp{Path: MustNewPointerFromString("/1")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(2))

			res, err = FindOp{Path: MustNewPointerFromString("/2")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(3))

			res, err = FindOp{Path: MustNewPointerFromString("/-1")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(3))
		})

		It("finds relative array item", func() {
			res, err := FindOp{Path: MustNewPointerFromString("/3:prev")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(3))
		})

		It("finds nested array item", func() {
			doc := []interface{}{[]interface{}{10, 11, 12}, 2, 3}

			res, err := FindOp{Path: MustNewPointerFromString("/0/1")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(11))
		})

		It("finds relative nested array item", func() {
			doc := []interface{}{1, []interface{}{10, 11, 12}, 3}

			res, err := FindOp{Path: MustNewPointerFromString("/0:next/1")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(11))
		})

		It("finds array item from an array that is inside a map", func() {
			doc := map[interface{}]interface{}{
				"abc": []interface{}{1, 2, 3},
			}

			res, err := FindOp{Path: MustNewPointerFromString("/abc/1")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(2))
		})

		It("returns an error if it's not an array when index is being accessed", func() {
			_, err := FindOp{Path: MustNewPointerFromString("/0")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/0' but found 'map[interface {}]interface {}'"))

			_, err = FindOp{Path: MustNewPointerFromString("/0/1")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/0' but found 'map[interface {}]interface {}'"))
		})

		It("returns an error if the index is out of bounds", func() {
			_, err := FindOp{Path: MustNewPointerFromString("/1")}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find array index '1' but found array of length '0' for path '/1'"))

			_, err = FindOp{Path: MustNewPointerFromString("/1/1")}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find array index '1' but found array of length '0' for path '/1'"))
		})
	})

	Describe("array with after last item", func() {
		It("returns an error as after-last-index tokens are not supported", func() {
			_, err := FindOp{Path: MustNewPointerFromString("/-")}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected not to find after last index token in path '/-' (not supported in find operations)"))
		})

		It("returns an error for nested array item", func() {
			doc := []interface{}{[]interface{}{10, 11, 12}, 2, 3}

			_, err := FindOp{Path: MustNewPointerFromString("/0/-")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected not to find after last index token in path '/0/-' (not supported in find operations)"))
		})

		It("returns an error array item from an array that is inside a map", func() {
			doc := map[interface{}]interface{}{
				"abc": []interface{}{1, 2, 3},
			}

			_, err := FindOp{Path: MustNewPointerFromString("/abc/-")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected not to find after last index token in path '/abc/-' (not supported in find operations)"))
		})

		It("returns an error if after last index token is not last", func() {
			ptr := NewPointer([]Token{RootToken{}, AfterLastIndexToken{}, KeyToken{}})

			_, err := FindOp{Path: ptr}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected not to find after last index token in path '/-/' (not supported in find operations)"))
		})

		It("returns an error if it's not an array being accessed", func() {
			_, err := FindOp{Path: MustNewPointerFromString("/-")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected not to find after last index token in path '/-' (not supported in find operations)"))

			doc := map[interface{}]interface{}{"key": map[interface{}]interface{}{}}

			_, err = FindOp{Path: MustNewPointerFromString("/key/-")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected not to find after last index token in path '/key/-' (not supported in find operations)"))
		})
	})

	Describe("array item with matching key and value", func() {
		It("finds array item if found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := FindOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"key": "val"}))
		})

		It("finds relative array item", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := FindOp{Path: MustNewPointerFromString("/key=val2:prev")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"key": "val"}))
		})

		It("returns an error if no items found and matching is not optional", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val2"},
				map[interface{}]interface{}{"key2": "val"},
			}

			_, err := FindOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find exactly one matching array item for path '/key=val' but found 0"))
		})

		It("returns an error if multiple items found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val"},
			}

			_, err := FindOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find exactly one matching array item for path '/key=val' but found 2"))
		})

		It("finds array item even if not all items are maps", func() {
			doc := []interface{}{
				3,
				map[interface{}]interface{}{"key": "val"},
			}

			res, err := FindOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"key": "val"}))
		})

		It("finds nested matching item", func() {
			doc := []interface{}{
				map[interface{}]interface{}{
					"key": "val",
					"items": []interface{}{
						map[interface{}]interface{}{"nested-key": "val"},
						map[interface{}]interface{}{"nested-key": "val2"},
					},
				},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := FindOp{Path: MustNewPointerFromString("/key=val/items/nested-key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"nested-key": "val"}))
		})

		It("finds relative nested matching item", func() {
			doc := []interface{}{
				map[interface{}]interface{}{
					"key": "val",
					"items": []interface{}{
						map[interface{}]interface{}{"nested-key": "val"},
						map[interface{}]interface{}{"nested-key": "val2"},
					},
				},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := FindOp{Path: MustNewPointerFromString("/key=val2:prev/items/nested-key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"nested-key": "val"}))
		})

		It("finds missing matching item if it does not exist", func() {
			doc := []interface{}{map[interface{}]interface{}{"xyz": "xyz"}}

			res, err := FindOp{Path: MustNewPointerFromString("/name=val?/efg")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("finds nested missing matching item if it does not exist", func() {
			doc := []interface{}{map[interface{}]interface{}{"xyz": "xyz"}}

			res, err := FindOp{Path: MustNewPointerFromString("/name=val?/efg/name=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"name": "val"}))
		})

		It("returns an error if it's not an array is being accessed", func() {
			_, err := FindOp{Path: MustNewPointerFromString("/key=val")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/key=val' but found 'map[interface {}]interface {}'"))

			_, err = FindOp{Path: MustNewPointerFromString("/key=val/items/key=val")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/key=val' but found 'map[interface {}]interface {}'"))
		})
	})

	Describe("map key", func() {
		It("finds map key", func() {
			doc := map[interface{}]interface{}{
				"abc": "abc",
				"xyz": "xyz",
			}

			res, err := FindOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("abc"))
		})

		It("finds nested map key", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": "efg",
					"opr": "opr",
				},
				"xyz": "xyz",
			}

			res, err := FindOp{Path: MustNewPointerFromString("/abc/efg")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("efg"))
		})

		It("finds nested map key that does not exist", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"opr": "opr"},
				"xyz": "xyz",
			}

			res, err := FindOp{Path: MustNewPointerFromString("/abc/efg?")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("finds super nested map key that does not exist", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": map[interface{}]interface{}{}, // wrong level
				},
			}

			res, err := FindOp{Path: MustNewPointerFromString("/abc/opr?/efg")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("returns an error if parent key does not exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			_, err := FindOp{Path: MustNewPointerFromString("/abc/efg")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found map keys: 'xyz')"))
		})

		It("returns an error if key does not exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz", 123: "xyz", "other-xyz": "xyz"}

			_, err := FindOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found map keys: 'other-xyz', 'xyz')"))
		})

		It("returns an error without other found keys when there are no keys and key does not exist", func() {
			doc := map[interface{}]interface{}{}

			_, err := FindOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found no other map keys)"))
		})

		It("returns nil for missing key if key is not expected to exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			res, err := FindOp{Path: MustNewPointerFromString("/abc?/efg")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("returns nil for nested missing keys if key is not expected to exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			res, err := FindOp{Path: MustNewPointerFromString("/abc?/other/efg")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("returns an error if missing key needs to be created but next access does not make sense", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			_, err := FindOp{Path: MustNewPointerFromString("/abc?/0")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find key or matching index token at path '/abc?/0'"))
		})

		It("returns an error if it's not a map when key is being accessed", func() {
			_, err := FindOp{Path: MustNewPointerFromString("/abc")}.Apply([]interface{}{1, 2, 3})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map at path '/abc' but found '[]interface {}'"))

			_, err = FindOp{Path: MustNewPointerFromString("/abc/efg")}.Apply([]interface{}{1, 2, 3})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map at path '/abc' but found '[]interface {}'"))
		})
	})
})
