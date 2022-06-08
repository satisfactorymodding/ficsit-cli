package cmd

import (
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply profiles to all installations",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO
		return nil
	},
}
