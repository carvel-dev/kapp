package patch_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch/patch"
)

type PointerTestCase struct {
	String string
	Tokens []Token
}

var testCases = []PointerTestCase{
	{"", []Token{RootToken{}}},

	// Root level
	{"/", []Token{RootToken{}, KeyToken{Key: ""}}},
	{"//", []Token{RootToken{}, KeyToken{Key: ""}, KeyToken{Key: ""}}},
	{"/ ", []Token{RootToken{}, KeyToken{Key: " "}}},

	// Maps
	{"/key", []Token{RootToken{}, KeyToken{Key: "key"}}},
	{"/key/", []Token{RootToken{}, KeyToken{Key: "key"}, KeyToken{Key: ""}}},
	{"/key/key2", []Token{RootToken{}, KeyToken{Key: "key"}, KeyToken{Key: "key2"}}},
	{"/key?/key2/key3", []Token{
		RootToken{},
		KeyToken{Key: "key", Optional: true},
		KeyToken{Key: "key2", Optional: true},
		KeyToken{Key: "key3", Optional: true},
	}},

	// Array indices
	{"/0", []Token{RootToken{}, IndexToken{Index: 0}}},
	{"/1000001", []Token{RootToken{}, IndexToken{Index: 1000001}}},
	{"/-2", []Token{RootToken{}, IndexToken{Index: -2}}},

	{"/-", []Token{RootToken{}, AfterLastIndexToken{}}},
	{"/ary/-", []Token{RootToken{}, KeyToken{Key: "ary"}, AfterLastIndexToken{}}},
	{"/-/key", []Token{RootToken{}, KeyToken{Key: "-"}, KeyToken{Key: "key"}}},

	{"/0:before", []Token{
		RootToken{},
		IndexToken{Index: 0, Modifiers: []Modifier{BeforeModifier{}}},
	}},
	{"/0:after", []Token{
		RootToken{},
		IndexToken{Index: 0, Modifiers: []Modifier{AfterModifier{}}},
	}},
	{"/-1:before", []Token{
		RootToken{},
		IndexToken{Index: -1, Modifiers: []Modifier{BeforeModifier{}}},
	}},
	{"/-1:prev:before", []Token{
		RootToken{},
		IndexToken{Index: -1, Modifiers: []Modifier{PrevModifier{}, BeforeModifier{}}},
	}},

	// Matching index token
	{"/name=val", []Token{RootToken{}, MatchingIndexToken{Key: "name", Value: "val"}}},
	{"/name=val?", []Token{RootToken{}, MatchingIndexToken{Key: "name", Value: "val", Optional: true}}},
	{"/name=val?/name2=val", []Token{
		RootToken{},
		MatchingIndexToken{Key: "name", Value: "val", Optional: true},
		MatchingIndexToken{Key: "name2", Value: "val", Optional: true},
	}},
	{"/=", []Token{RootToken{}, MatchingIndexToken{Key: "", Value: ""}}},
	{"/=?", []Token{RootToken{}, MatchingIndexToken{Key: "", Value: "", Optional: true}}},
	{"/name=", []Token{RootToken{}, MatchingIndexToken{Key: "name", Value: ""}}},
	{"/=val", []Token{RootToken{}, MatchingIndexToken{Key: "", Value: "val"}}},
	{"/==", []Token{RootToken{}, MatchingIndexToken{Key: "", Value: "="}}},

	{"/name=val:before", []Token{
		RootToken{},
		MatchingIndexToken{Key: "name", Value: "val", Modifiers: []Modifier{BeforeModifier{}}},
	}},
	{"/name=val:after", []Token{
		RootToken{},
		MatchingIndexToken{Key: "name", Value: "val", Modifiers: []Modifier{AfterModifier{}}},
	}},

	// Optionality
	{"/key?/name=val", []Token{
		RootToken{},
		KeyToken{Key: "key", Optional: true},
		MatchingIndexToken{Key: "name", Value: "val", Optional: true},
	}},
	{"/name=val?/key", []Token{
		RootToken{},
		MatchingIndexToken{Key: "name", Value: "val", Optional: true},
		KeyToken{Key: "key", Optional: true},
	}},

	// Escaping (todo support ~2 for '?'; ~3 for '=')
	{"/m~0n", []Token{RootToken{}, KeyToken{Key: "m~n"}}},
	{"/a~01b", []Token{RootToken{}, KeyToken{Key: "a~1b"}}},
	{"/a~1b", []Token{RootToken{}, KeyToken{Key: "a/b"}}},
	{"/name~0n=val~0n", []Token{RootToken{}, MatchingIndexToken{Key: "name~n", Value: "val~n"}}},
	{"/m~7n", []Token{RootToken{}, KeyToken{Key: "m:n"}}},

	// Special chars
	{"/c%d", []Token{RootToken{}, KeyToken{Key: "c%d"}}},
	{"/e^f", []Token{RootToken{}, KeyToken{Key: "e^f"}}},
	{"/g|h", []Token{RootToken{}, KeyToken{Key: "g|h"}}},
	{"/i\\j", []Token{RootToken{}, KeyToken{Key: "i\\j"}}},
	{"/k\"l", []Token{RootToken{}, KeyToken{Key: "k\"l"}}},
}

var _ = Describe("NewPointer", func() {
	It("panics if no tokens are given", func() {
		Expect(func() { NewPointer([]Token{}) }).To(Panic())
	})

	It("panics if first token is not root token", func() {
		Expect(func() { NewPointer([]Token{IndexToken{}}) }).To(Panic())
	})

	It("succeeds for basic case", func() {
		Expect(NewPointer([]Token{RootToken{}}).Tokens()).To(Equal([]Token{RootToken{}}))
	})
})

var _ = Describe("NewPointerFromString", func() {
	It("returns error if string doesn't start with /", func() {
		_, err := NewPointerFromString("abc")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to start with '/'"))
	})

	It("returns error if string includes unknown modifiers", func() {
		_, err := NewPointerFromString("/abc:unknown")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to find one of the following modifiers: 'prev', 'next', 'before', or 'after' but found 'unknown'"))
	})

	It("returns error if string has modifiers in after-last-index-token", func() {
		_, err := NewPointerFromString("/-:prev")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected not to find any modifiers with after last index token"))
	})

	It("returns error if string has modifiers in key-token", func() {
		_, err := NewPointerFromString("/key:prev")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected not to find any modifiers with key token"))
	})
})

var _ = Describe("Pointer.String", func() {
	for _, tc := range testCases {
		tc := tc // copy
		It(fmt.Sprintf("'%#v' results in '%s'", tc.Tokens, tc.String), func() {
			Expect(NewPointer(tc.Tokens).String()).To(Equal(tc.String))
		})
	}
})

var _ = Describe("Pointer.Tokens", func() {
	parsingTestCases := []PointerTestCase{
		{"/key/key2?", []Token{
			RootToken{},
			KeyToken{Key: "key"},
			KeyToken{Key: "key2", Optional: true},
		}},
	}

	parsingTestCases = append(parsingTestCases, testCases...)

	for _, tc := range parsingTestCases {
		tc := tc // copy
		It(fmt.Sprintf("'%s' results in '%#v'", tc.String, tc.Tokens), func() {
			Expect(MustNewPointerFromString(tc.String).Tokens()).To(Equal(tc.Tokens))
		})
	}
})

var _ = Describe("Pointer.IsSet", func() {
	It("returns true if there is at least one token", func() {
		Expect(MustNewPointerFromString("").IsSet()).To(BeTrue())
	})

	It("returns false if it's not a valid pointer", func() {
		Expect(Pointer{}.IsSet()).To(BeFalse())
	})
})

var _ = Describe("Pointer.UnmarshalFlag", func() {
	It("parses pointer if it's valid", func() {
		ptr := &Pointer{}

		err := ptr.UnmarshalFlag("/abc")
		Expect(err).ToNot(HaveOccurred())
		Expect(*ptr).To(Equal(MustNewPointerFromString("/abc")))
	})

	It("returns error if string doesn't start with /", func() {
		ptr := &Pointer{}

		err := ptr.UnmarshalFlag("abc")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to start with '/'"))
	})
})
