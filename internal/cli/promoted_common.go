// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCommonPromotedCmd(flags *rootFlags) *cobra.Command {
	var bodyUrl string

	cmd := &cobra.Command{
		Use:         "common",
		Short:       "Get temporary download URL for generated file",
		Long:        "Shortcut for 'common download-url'. Get temporary download URL for generated file",
		Example:     "  kie-cli common --url https://example.com/resource",
		Annotations: map[string]string{"pp:endpoint": "common.download-url", "pp:method": "POST", "pp:path": "/api/v1/common/download-url"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("url") && !flags.dryRun {
				return fmt.Errorf("required flag \"%s\" not set", "url")
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			path := "/api/v1/common/download-url"
			params := map[string]string{}
			// HasStore + non-GET falls through to a live API call here
			// rather than through resolveRead (GET-only internally); a
			// body-aware cached read helper is filed as #425 for when a
			// second store-backed POST-search consumer ships.
			body := map[string]any{}
			if bodyUrl != "" {
				body["url"] = bodyUrl
			}
			data, _, err := c.PostWithParams(path, params, body)

			if err != nil {
				return classifyAPIError(err, flags)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				var items []map[string]any
				if json.Unmarshal(data, &items) == nil && len(items) > 0 {
					if err := printAutoTable(cmd.OutOrStdout(), items); err != nil {
						return err
					}
					if len(items) >= 25 {
						fmt.Fprintf(os.Stderr, "\nShowing %d results. To narrow: add --limit, --json --select, or filter flags.\n", len(items))
					}
					return nil
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().StringVar(&bodyUrl, "url", "", "Url")

	// Wire sibling endpoints and sub-resources as subcommands

	return cmd
}
