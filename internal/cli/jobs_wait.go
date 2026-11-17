// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.
package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newJobsWaitCmd(flags *rootFlags) *cobra.Command {
	var taskID string
	var interval time.Duration
	var maxWait time.Duration

	cmd := &cobra.Command{
		Use:     "wait",
		Short:   "Poll task until success, failure, or timeout",
		Example: "  kie-cli jobs wait --task-id TASK_ID --agent --select data.state,data.resultJson",
		Annotations: map[string]string{
			"pp:workflow":   "jobs.wait",
			"mcp:read-only": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if taskID == "" && !flags.dryRun {
				return fmt.Errorf("required flag \"task-id\" not set")
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			deadline := time.Now().Add(maxWait)
			var last []byte
			for {
				if time.Now().After(deadline) {
					return fmt.Errorf("timeout waiting for task %s", taskID)
				}
				params := map[string]string{"taskId": taskID}
				last, err = c.Get("/api/v1/jobs/recordInfo", params)
				if err != nil {
					return classifyAPIError(err, flags)
				}
				var envelope map[string]any
				if err := json.Unmarshal(last, &envelope); err != nil {
					return fmt.Errorf("parse recordInfo: %w", err)
				}
				data, _ := envelope["data"].(map[string]any)
				state, _ := data["state"].(string)
				switch state {
				case "success":
					return printOutputWithFlags(cmd.OutOrStdout(), last, flags)
				case "fail", "failed", "error":
					return fmt.Errorf("task failed: %s", string(last))
				}
				if flags.dryRun {
					return printOutputWithFlags(cmd.OutOrStdout(), last, flags)
				}
				time.Sleep(interval)
			}
		},
	}
	cmd.Flags().StringVar(&taskID, "task-id", "", "Task id from create-task")
	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Poll interval")
	cmd.Flags().DurationVar(&maxWait, "max-wait", 10*time.Minute, "Stop polling after this duration")
	return cmd
}
