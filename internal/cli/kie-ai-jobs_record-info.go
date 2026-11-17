// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newKieAiJobsRecordInfoCmd(flags *rootFlags) *cobra.Command {
	var flagTaskId string

	cmd := &cobra.Command{
		Use:         "record-info",
		Aliases:     []string{"list"},
		Short:       "Get task status and result",
		Example:     "  kie-cli kie-ai-jobs record-info --task-id 550e8400-e29b-41d4-a716-446655440000",
		Annotations: map[string]string{"pp:endpoint": "kie-ai-jobs.record-info", "pp:method": "GET", "pp:path": "/api/v1/jobs/recordInfo", "mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("task-id") && !flags.dryRun {
				return fmt.Errorf("required flag \"%s\" not set", "task-id")
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			path := "/api/v1/jobs/recordInfo"
			params := map[string]string{}
			if flagTaskId != "" {
				params["taskId"] = fmt.Sprintf("%v", flagTaskId)
			}
			data, err := c.Get(path, params)
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
	cmd.Flags().StringVar(&flagTaskId, "task-id", "", "Task id")

	return cmd
}
