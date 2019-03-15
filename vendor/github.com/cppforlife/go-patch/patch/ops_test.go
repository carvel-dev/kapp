package patch_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch/patch"
)

var _ = Describe("Ops.Apply", func() {
	It("runs through all operations", func() {
		ops := Ops([]Op{
			RemoveOp{Path: MustNewPointerFromString("/0")},
			RemoveOp{Path: MustNewPointerFromString("/0")},
		})

		res, err := ops.Apply([]interface{}{1, 2, 3})
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal([]interface{}{3}))
	})

	It("returns original input if there are no operations", func() {
		res, err := Ops([]Op{}).Apply([]interface{}{1, 2, 3})
		Expect(err).ToNot(HaveOccurred())
		Expect(res).To(Equal([]interface{}{1, 2, 3}))
	})

	It("returns error if any operation errors", func() {
		ops := Ops([]Op{
			RemoveOp{Path: MustNewPointerFromString("/0")},
			ErrOp{errors.New("fake-err")},
		})

		_, err := ops.Apply([]interface{}{1, 2, 3})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-err"))
	})
})
