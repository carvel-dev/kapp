package patch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch/patch"
)

var _ = Describe("RemoveOp.Apply", func() {
	It("returns an error if path is for the entire document", func() {
		_, err := RemoveOp{Path: MustNewPointerFromString("")}.Apply("a")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Cannot remove entire document"))
	})

	Describe("array item", func() {
		It("removes array item", func() {
			res, err := RemoveOp{Path: MustNewPointerFromString("/0")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{2, 3}))

			res, err = RemoveOp{Path: MustNewPointerFromString("/1")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 3}))

			res, err = RemoveOp{Path: MustNewPointerFromString("/2")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2}))

			res, err = RemoveOp{Path: MustNewPointerFromString("/-1")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2}))

			res, err = RemoveOp{Path: MustNewPointerFromString("/0")}.Apply([]interface{}{1})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{}))
		})

		It("removes relative array item", func() {
			res, err := RemoveOp{Path: MustNewPointerFromString("/3:prev")}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2}))
		})

		It("removes nested array item", func() {
			res, err := RemoveOp{Path: MustNewPointerFromString("/0/1")}.Apply([]interface{}{[]interface{}{10, 11, 12}, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{[]interface{}{10, 12}, 2, 3}))
		})

		It("removes relative nested array item", func() {
			res, err := RemoveOp{Path: MustNewPointerFromString("/1:prev/1")}.Apply([]interface{}{[]interface{}{10, 11, 12}, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{[]interface{}{10, 12}, 2, 3}))
		})

		It("removes array item from an array that is inside a map", func() {
			doc := map[interface{}]interface{}{
				"abc": []interface{}{1, 2, 3},
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/abc/1")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": []interface{}{1, 3},
			}))
		})

		It("returns an error if it's not an array when index is being accessed", func() {
			_, err := RemoveOp{Path: MustNewPointerFromString("/0")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/0' but found 'map[interface {}]interface {}'"))

			_, err = RemoveOp{Path: MustNewPointerFromString("/0/1")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/0' but found 'map[interface {}]interface {}'"))
		})

		It("returns an error if the index is out of bounds", func() {
			_, err := RemoveOp{Path: MustNewPointerFromString("/1")}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find array index '1' but found array of length '0' for path '/1'"))

			_, err = RemoveOp{Path: MustNewPointerFromString("/1/1")}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find array index '1' but found array of length '0' for path '/1'"))
		})
	})

	It("returns an error if after last token is found", func() {
		_, err := RemoveOp{Path: MustNewPointerFromString("/-")}.Apply([]interface{}{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(
			"Expected to not find token 'patch.AfterLastIndexToken' at path '/-'"))
	})

	Describe("array item with matching key and value", func() {
		It("removes array item if found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{"key": "val2"},
			}))
		})

		It("removes array item if found, leaving empty array", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{}))
		})

		It("removes relative array item", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/key=val:next")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{"key": "val"},
			}))
		})

		It("returns an error if no items found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val2"},
				map[interface{}]interface{}{"key2": "val"},
			}

			_, err := RemoveOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find exactly one matching array item for path '/key=val' but found 0"))
		})

		It("returns an error if multiple items found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val"},
			}

			_, err := RemoveOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find exactly one matching array item for path '/key=val' but found 2"))
		})

		It("removes array item even if not all items are maps", func() {
			doc := []interface{}{
				3,
				map[interface{}]interface{}{"key": "val"},
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{3}))
		})

		It("removes nested matching item", func() {
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

			res, err := RemoveOp{Path: MustNewPointerFromString("/key=val/items/nested-key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{
					"key": "val",
					"items": []interface{}{
						map[interface{}]interface{}{"nested-key": "val2"},
					},
				},
				map[interface{}]interface{}{"key": "val2"},
			}))
		})

		It("removes relative nested matching item", func() {
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

			res, err := RemoveOp{Path: MustNewPointerFromString("/key=val2:prev/items/nested-key=val")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{
					"key": "val",
					"items": []interface{}{
						map[interface{}]interface{}{"nested-key": "val2"},
					},
				},
				map[interface{}]interface{}{"key": "val2"},
			}))
		})

		It("removes nested matching item that does not exist", func() {
			doc := map[interface{}]interface{}{
				"abc": []interface{}{
					map[interface{}]interface{}{"opr": "opr"},
				},
				"xyz": "xyz",
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/abc/opr=not-opr?")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": []interface{}{
					map[interface{}]interface{}{"opr": "opr"},
				},
				"xyz": "xyz",
			}))
		})

		It("returns an error if it's not an array is being accessed", func() {
			_, err := RemoveOp{Path: MustNewPointerFromString("/key=val")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/key=val' but found 'map[interface {}]interface {}'"))

			_, err = RemoveOp{Path: MustNewPointerFromString("/key=val/items/key=val")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/key=val' but found 'map[interface {}]interface {}'"))
		})
	})

	Describe("map key", func() {
		It("removes map key", func() {
			doc := map[interface{}]interface{}{
				"abc": "abc",
				"xyz": "xyz",
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"xyz": "xyz"}))
		})

		It("removes nested map key", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": "efg",
					"opr": "opr",
				},
				"xyz": "xyz",
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/abc/efg")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"opr": "opr"},
				"xyz": "xyz",
			}))
		})

		It("removes nested map key that does not exist", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"opr": "opr"},
				"xyz": "xyz",
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/abc/efg?")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"opr": "opr"},
				"xyz": "xyz",
			}))
		})

		It("removes super nested map key that does not exist", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": map[interface{}]interface{}{}, // wrong level
				},
			}

			res, err := RemoveOp{Path: MustNewPointerFromString("/abc/opr?/efg")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": map[interface{}]interface{}{}, // wrong level
				},
			}))
		})

		It("returns an error if parent key does not exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			_, err := RemoveOp{Path: MustNewPointerFromString("/abc/efg")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found map keys: 'xyz')"))
		})

		It("returns an error if key does not exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz", 123: "xyz", "other-xyz": "xyz"}

			_, err := RemoveOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found map keys: 'other-xyz', 'xyz')"))
		})

		It("returns an error without other found keys when there are no keys and key does not exist", func() {
			doc := map[interface{}]interface{}{}

			_, err := RemoveOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found no other map keys)"))
		})

		It("returns an error if it's not a map when key is being accessed", func() {
			_, err := RemoveOp{Path: MustNewPointerFromString("/abc")}.Apply([]interface{}{1, 2, 3})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map at path '/abc' but found '[]interface {}'"))

			_, err = RemoveOp{Path: MustNewPointerFromString("/abc/efg")}.Apply([]interface{}{1, 2, 3})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map at path '/abc' but found '[]interface {}'"))
		})
	})
})
