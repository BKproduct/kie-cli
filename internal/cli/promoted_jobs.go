// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.
package cli

import "github.com/spf13/cobra"

func newJobsPromotedCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "KIE async jobs — create, poll, wait",
		Long:  "Unified job API for KIE market models (nano-banana-pro, sora-2, kling, etc.).",
	}
	cmd.AddCommand(newKieAiJobsCreateTaskCmd(flags))
	cmd.AddCommand(newKieAiJobsRecordInfoCmd(flags))
	cmd.AddCommand(newJobsWaitCmd(flags))
	return cmd
}
