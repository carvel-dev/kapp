package patch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	. "github.com/cppforlife/go-patch/patch"
)

var _ = Describe("YAML compatibility", func() {
	Describe("empty string", func() {
		It("[workaround] works deserializing empty strings", func() {
			str := `
- type: replace
  path: /instance_groups/name=cloud_controller/instances
  value: !!str ""
`

			var opDefs []OpDefinition

			err := yaml.Unmarshal([]byte(str), &opDefs)
			Expect(err).ToNot(HaveOccurred())

			val := opDefs[0].Value
			Expect((*val).(string)).To(Equal(""))
		})

		It("[fixed] does not work deserializing empty strings", func() {
			str := `
- type: replace
  path: /instance_groups/name=cloud_controller/instances
  value: ""
`

			var opDefs []OpDefinition

			err := yaml.Unmarshal([]byte(str), &opDefs)
			Expect(err).ToNot(HaveOccurred())

			val := opDefs[0].Value
			Expect((*val).(string)).To(Equal(""))
		})
	})
})
