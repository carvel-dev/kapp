/*
 * Copyright 2020 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"encoding/json"

	"github.com/k14s/kapp/pkg/kapp/website"
	"github.com/spf13/cobra"
)

type WebsiteOptions struct {
	ListenAddr string
}

func NewWebsiteOptions() *WebsiteOptions {
	return &WebsiteOptions{}
}

func NewWebsiteCmd(o *WebsiteOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "website",
		Hidden: true,
		Short:  "Starts website HTTP server",
		RunE:   func(_ *cobra.Command, _ []string) error { return o.Run() },
	}
	cmd.Flags().StringVar(&o.ListenAddr, "listen-addr", "localhost:8080", "Listen address")
	return cmd
}

func (o *WebsiteOptions) Server() *website.Server {
	opts := website.ServerOpts{
		ListenAddr: o.ListenAddr,
		ErrorFunc: func(err error) ([]byte, error) {
			return json.Marshal(map[string]string{"error": err.Error()})
		},
	}
	return website.NewServer(opts)
}

func (o *WebsiteOptions) Run() error {
	return o.Server().Run()
}
