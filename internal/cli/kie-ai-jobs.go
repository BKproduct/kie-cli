// Copyright 2026 boris-korol. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"github.com/spf13/cobra"
)

func newKieAiJobsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "kie-ai-jobs",
		Short:  "Manage kie ai jobs",
		Hidden: true,
		RunE:   parentNoSubcommandRunE(flags),
	}

	cmd.AddCommand(newKieAiJobsCreateTaskCmd(flags))
	cmd.AddCommand(newKieAiJobsRecordInfoCmd(flags))
	return cmd
}
