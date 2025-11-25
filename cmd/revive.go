package cmd

import (
	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var reviveCmd = &cobra.Command{
	Use:   "revive",
	Short: "Attempt reviving an irresponsive container",
	Long:  "Attempt reviving an irresponsive container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cli, ctx, err := docker.GetClient()
		if err != nil {
			return err
		}
		if err = docker.StartContainer(cli, ctx, args[0]); err != nil {
			return err
		}
		return nil
	},
}
