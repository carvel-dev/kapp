package patch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch/patch"
)

var _ = Describe("TestOp.Apply", func() {
	Describe("value check", func() {
		It("does not error and returns original document if value matches", func() {
			res, err := TestOp{
				Path:  MustNewPointerFromString("/0"),
				Value: 1,
			}.Apply([]interface{}{1, 2, 3})

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{1, 2, 3}))

			res, err = TestOp{
				Path:  MustNewPointerFromString("/0"),
				Value: nil,
			}.Apply([]interface{}{nil, 2, 3})

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{nil, 2, 3}))
		})

		It("returns an error if value does not match", func() {
			_, err := TestOp{
				Path:  MustNewPointerFromString("/0"),
				Value: 2,
			}.Apply([]interface{}{1, 2, 3})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Found value does not match expected value"))

			_, err = TestOp{
				Path:  MustNewPointerFromString("/0"),
				Value: 2,
			}.Apply([]interface{}{nil})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Found value does not match expected value"))
		})
	})

	Describe("absence check", func() {
		It("does not error and returns original document if key is absent", func() {
			res, err := TestOp{
				Path:   MustNewPointerFromString("/0"),
				Absent: true,
			}.Apply([]interface{}{})

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]interface{}{}))

			res, err = TestOp{
				Path:   MustNewPointerFromString("/a"),
				Absent: true,
			}.Apply(map[interface{}]interface{}{"b": 123})

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(map[interface{}]interface{}{"b": 123}))
		})

		It("returns an error if parent key is absent", func() {
			_, err := TestOp{
				Path:   MustNewPointerFromString("/0/0"),
				Absent: true,
			}.Apply([]interface{}{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find array index '0' but found array of length '0' for path '/0'"))

			_, err = TestOp{
				Path:   MustNewPointerFromString("/a/b"),
				Absent: true,
			}.Apply(map[interface{}]interface{}{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to find a map key 'a' for path '/a' (found no other map keys)"))
		})

		It("returns an error if key is present", func() {
			_, err := TestOp{
				Path:   MustNewPointerFromString("/0"),
				Absent: true,
			}.Apply([]interface{}{1})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to not find '/0'"))

			_, err = TestOp{
				Path:   MustNewPointerFromString("/a"),
				Absent: true,
			}.Apply(map[interface{}]interface{}{"a": 123})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected to not find '/a'"))
		})
	})
})
