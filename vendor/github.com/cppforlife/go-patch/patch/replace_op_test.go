package patch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch/patch"
)

var _ = Describe("ReplaceOp.Apply", func() {
	It("returns error if replacement value cloning fails", func() {
		_, err := ReplaceOp{Path: MustNewPointerFromString(""), Value: func() {}}.Apply("a")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("ReplaceOp cloning value"))
	})

	It("uses cloned value for replacement", func() {
		repVal := map[interface{}]interface{}{"a": "b"}

		res, err := ReplaceOp{Path: MustNewPointerFromString(""), Value: repVal}.Apply("a")
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal(repVal))

		res.(map[interface{}]interface{})["c"] = "d"
		Expect(res).ToNot(Equal(repVal))
	})

	It("replaces document if path is for the entire document", func() {
		res, err := ReplaceOp{Path: MustNewPointerFromString(""), Value: "b"}.Apply("a")
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal("b"))
	})

	Describe("array item", func() {
		It("replaces array item", func() {
			res, err := ReplaceOp{Path: MustNewPointerFromString("/0"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{10, 2, 3}))

			res, err = ReplaceOp{Path: MustNewPointerFromString("/1"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 10, 3}))

			res, err = ReplaceOp{Path: MustNewPointerFromString("/2"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2, 10}))
		})

		It("replaces relative array item", func() {
			res, err := ReplaceOp{Path: MustNewPointerFromString("/3:prev"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2, 10}))

			res, err = ReplaceOp{Path: MustNewPointerFromString("/0:before"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{10, 1, 2, 3}))

			res, err = ReplaceOp{Path: MustNewPointerFromString("/1:before"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 10, 2, 3}))

			res, err = ReplaceOp{Path: MustNewPointerFromString("/3:prev:after"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2, 3, 10}))
		})

		It("replaces nested array item", func() {
			doc := []interface{}{[]interface{}{10, 11, 12}, 2, 3}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/0/1"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{[]interface{}{10, 100, 12}, 2, 3}))
		})

		It("replaces relative nested array item", func() {
			doc := []interface{}{1, []interface{}{10, 11, 12}, 3}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/0:next/1"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, []interface{}{10, 100, 12}, 3}))
		})

		It("replaces array item from an array that is inside a map", func() {
			doc := map[interface{}]interface{}{
				"abc": []interface{}{1, 2, 3},
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc/1"), Value: 10}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": []interface{}{1, 10, 3},
			}))
		})

		It("returns an error if it's not an array when index is being accessed", func() {
			_, err := ReplaceOp{Path: MustNewPointerFromString("/0")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/0' but found 'map[interface {}]interface {}'"))

			_, err = ReplaceOp{Path: MustNewPointerFromString("/0/1")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/0' but found 'map[interface {}]interface {}'"))
		})

		It("returns an error if the index is out of bounds", func() {
			_, err := ReplaceOp{Path: MustNewPointerFromString("/1")}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find array index '1' but found array of length '0' for path '/1'"))

			_, err = ReplaceOp{Path: MustNewPointerFromString("/1/1")}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find array index '1' but found array of length '0' for path '/1'"))
		})
	})

	Describe("array with after last item", func() {
		It("appends new item", func() {
			res, err := ReplaceOp{Path: MustNewPointerFromString("/-"), Value: 10}.Apply([]interface{}{})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{10}))

			res, err = ReplaceOp{Path: MustNewPointerFromString("/-"), Value: 10}.Apply([]interface{}{1, 2, 3})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2, 3, 10}))
		})

		It("appends nested array item", func() {
			doc := []interface{}{[]interface{}{10, 11, 12}, 2, 3}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/0/-"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{[]interface{}{10, 11, 12, 100}, 2, 3}))
		})

		It("appends array item from an array that is inside a map", func() {
			doc := map[interface{}]interface{}{
				"abc": []interface{}{1, 2, 3},
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc/-"), Value: 10}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": []interface{}{1, 2, 3, 10},
			}))
		})

		It("returns an error if after last index token is not last", func() {
			ptr := NewPointer([]Token{RootToken{}, AfterLastIndexToken{}, KeyToken{}})

			_, err := ReplaceOp{Path: ptr}.Apply([]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected after last index token to be last in path '/-/'"))
		})

		It("returns an error if it's not an array being accessed", func() {
			_, err := ReplaceOp{Path: MustNewPointerFromString("/-")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/-' but found 'map[interface {}]interface {}'"))

			doc := map[interface{}]interface{}{"key": map[interface{}]interface{}{}}

			_, err = ReplaceOp{Path: MustNewPointerFromString("/key/-")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/key/-' but found 'map[interface {}]interface {}'"))
		})
	})

	Describe("array item with matching key and value", func() {
		It("replaces array item if found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/key=val"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{
				100,
				map[interface{}]interface{}{"key": "val2"},
			}))
		})

		It("replaces relative array item if found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/key=val2:prev"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{
				100,
				map[interface{}]interface{}{"key": "val2"},
			}))

			doc = []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err = ReplaceOp{Path: MustNewPointerFromString("/key=val2:before"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{"key": "val"},
				100,
				map[interface{}]interface{}{"key": "val2"},
			}))

			doc = []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
			}

			res, err = ReplaceOp{Path: MustNewPointerFromString("/key=val:next:after"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val2"},
				100,
			}))
		})

		It("returns an error if no items found and matching is not optional", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val2"},
				map[interface{}]interface{}{"key2": "val"},
			}

			_, err := ReplaceOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find exactly one matching array item for path '/key=val' but found 0"))
		})

		It("returns an error if multiple items found", func() {
			doc := []interface{}{
				map[interface{}]interface{}{"key": "val"},
				map[interface{}]interface{}{"key": "val"},
			}

			_, err := ReplaceOp{Path: MustNewPointerFromString("/key=val")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find exactly one matching array item for path '/key=val' but found 2"))
		})

		It("replaces array item even if not all items are maps", func() {
			doc := []interface{}{
				3,
				map[interface{}]interface{}{"key": "val"},
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/key=val"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{3, 100}))
		})

		It("replaces nested matching item", func() {
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

			res, err := ReplaceOp{Path: MustNewPointerFromString("/key=val/items/nested-key=val"), Value: 100}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{
					"key": "val",
					"items": []interface{}{
						100,
						map[interface{}]interface{}{"nested-key": "val2"},
					},
				},
				map[interface{}]interface{}{"key": "val2"},
			}))
		})

		It("appends missing matching item if it does not exist", func() {
			doc := []interface{}{map[interface{}]interface{}{"xyz": "xyz"}}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/name=val?/efg"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{"xyz": "xyz"},
				map[interface{}]interface{}{
					"name": "val",
					"efg":  1,
				},
			}))
		})

		It("appends nested missing matching item if it does not exist", func() {
			doc := []interface{}{map[interface{}]interface{}{"xyz": "xyz"}}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/name=val?/efg/name=val"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal([]interface{}{
				map[interface{}]interface{}{"xyz": "xyz"},
				map[interface{}]interface{}{
					"name": "val",
					"efg":  []interface{}{1},
				},
			}))
		})

		It("returns an error if it's not an array is being accessed", func() {
			_, err := ReplaceOp{Path: MustNewPointerFromString("/key=val")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/key=val' but found 'map[interface {}]interface {}'"))

			_, err = ReplaceOp{Path: MustNewPointerFromString("/key=val/items/key=val")}.Apply(map[interface{}]interface{}{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find an array at path '/key=val' but found 'map[interface {}]interface {}'"))
		})
	})

	Describe("map key", func() {
		It("replaces map key", func() {
			doc := map[interface{}]interface{}{
				"abc": "abc",
				"xyz": "xyz",
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"abc": 1, "xyz": "xyz"}))

			res, err = ReplaceOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"abc": nil, "xyz": "xyz"}))
		})

		It("replaces nested map key", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": "efg",
					"opr": "opr",
				},
				"xyz": "xyz",
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc/efg"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"efg": 1, "opr": "opr"},
				"xyz": "xyz",
			}))
		})

		It("replaces nested map key that does not exist", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"opr": "opr"},
				"xyz": "xyz",
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc/efg?"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"efg": 1, "opr": "opr"},
				"xyz": "xyz",
			}))
		})

		It("replaces super nested map key that does not exist", func() {
			doc := map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": map[interface{}]interface{}{}, // wrong level
				},
			}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc/opr?/efg"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"efg": map[interface{}]interface{}{}, // wrong level
					"opr": map[interface{}]interface{}{"efg": 1},
				},
			}))
		})

		It("returns an error if parent key does not exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			_, err := ReplaceOp{Path: MustNewPointerFromString("/abc/efg")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found map keys: 'xyz')"))
		})

		It("returns an error if key does not exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz", 123: "xyz", "other-xyz": "xyz"}

			_, err := ReplaceOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found map keys: 'other-xyz', 'xyz')"))
		})

		It("returns an error without other found keys when there are no keys and key does not exist", func() {
			doc := map[interface{}]interface{}{}

			_, err := ReplaceOp{Path: MustNewPointerFromString("/abc")}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map key 'abc' for path '/abc' (found no other map keys)"))
		})

		It("creates missing key if key is not expected to exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc?/efg"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{"efg": 1},
				"xyz": "xyz",
			}))
		})

		It("creates nested missing keys if key is not expected to exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc?/other/efg"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": map[interface{}]interface{}{
					"other": map[interface{}]interface{}{"efg": 1},
				},
				"xyz": "xyz",
			}))
		})

		It("creates missing key with array value for index access if key is not expected to exist", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			res, err := ReplaceOp{Path: MustNewPointerFromString("/abc?/-"), Value: 1}.Apply(doc)
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(Equal(map[interface{}]interface{}{
				"abc": []interface{}{1},
				"xyz": "xyz",
			}))
		})

		It("returns an error if missing key needs to be created but next access does not make sense", func() {
			doc := map[interface{}]interface{}{"xyz": "xyz"}

			_, err := ReplaceOp{Path: MustNewPointerFromString("/abc?/0"), Value: 1}.Apply(doc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find key, matching index or after last index token at path '/abc?/0'"))
		})

		It("returns an error if it's not a map when key is being accessed", func() {
			_, err := ReplaceOp{Path: MustNewPointerFromString("/abc")}.Apply([]interface{}{1, 2, 3})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map at path '/abc' but found '[]interface {}'"))

			_, err = ReplaceOp{Path: MustNewPointerFromString("/abc/efg")}.Apply([]interface{}{1, 2, 3})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected to find a map at path '/abc' but found '[]interface {}'"))
		})
	})
})
