// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package preflight

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	ctldgraph "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diffgraph"
)

const preflightFlag = "preflight"

// Registry is a collection of preflight checks
type Registry struct {
	known map[string]Check
}

// NewRegistry will return a new *Registry with the
// provided set of preflight checks added to the registry
func NewRegistry(checks map[string]Check) *Registry {
	registry := &Registry{}
	for name, check := range checks {
		registry.AddCheck(name, check)
	}
	return registry
}

// String returns a string representation of the
// enabled preflight checks. It follows the format:
// CheckName,...
// This method is needed so Registry implements
// the pflag.Value interface
func (c *Registry) String() string {
	enabled := []string{}
	for k, v := range c.known {
		if v.Enabled() {
			enabled = append(enabled, k)
		}
	}
	return strings.Join(enabled, ",")
}

// Type returns a string representing the type
// of the Registry. It is needed to implement the
// pflag.Value interface
func (c *Registry) Type() string {
	return fmt.Sprintf("%T", c)
}

// Set takes in a string in the format of
// CheckName,...
// and sets the specified preflight check
// as enabled if listed, otherwise, sets as
// disabled if not listed.
// Returns an error if there is a problem
// parsing the preflight checks
func (c *Registry) Set(s string) error {
	if c.known == nil {
		return nil
	}

	enabled := map[string]struct{}{}
	// enable those specified
	mappings := strings.Split(s, ",")
	for _, key := range mappings {
		if _, ok := c.known[key]; !ok {
			return fmt.Errorf("unknown preflight check %q specified", key)
		}
		c.known[key].SetEnabled(true)
		enabled[key] = struct{}{}
	}
	// disable unspecified validators
	for key := range c.known {
		if _, ok := enabled[key]; !ok {
			c.known[key].SetEnabled(false)
		}
	}
	return nil
}

// AddFlags adds the --preflight flag to a
// pflag.FlagSet and configures the preflight
// checks in the registry based on the user provided
// values. If no values are provided by a user the
// default values are used.
func (c *Registry) AddFlags(flags *pflag.FlagSet) {
	knownChecks := []string{}
	for name := range c.known {
		knownChecks = append(knownChecks, name)
	}
	flags.Var(c, preflightFlag, fmt.Sprintf("preflight checks to run. Available preflight checks are [%s]", strings.Join(knownChecks, ",")))
}

// AddCheck adds a new preflight check to the registry.
// The name provided will map to the provided Check.
func (c *Registry) AddCheck(name string, check Check) {
	if c.known == nil {
		c.known = make(map[string]Check)
	}
	c.known[name] = check
}

// Run will execute any enabled preflight checks. The provided
// Context and ChangeGraph will be passed to the preflight checks
// that are being executed.
func (c *Registry) Run(ctx context.Context, cg *ctldgraph.ChangeGraph) error {
	for name, check := range c.known {
		if check.Enabled() {
			err := check.Run(ctx, cg)
			if err != nil {
				return fmt.Errorf("running preflight check %q: %w", name, err)
			}
		}
	}
	return nil
}
