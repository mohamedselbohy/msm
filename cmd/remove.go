package cmd

import (
	"fmt"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "rm",
	Short: "Delete a workspace",
	Long:  "Delete a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cli, ctx, err := docker.GetClient()
		workspace := args[0]
		if err != nil {
			return err
		}
		if exists, err := docker.SearchRunningContainers("ros-" + workspace); err == nil {
			if !exists {
				return fmt.Errorf("error: Workspace is not active")
			}
		} else {
			return fmt.Errorf("error: Failed to search for containers")
		}

		if err = docker.StopAndDeleteContainer(cli, ctx, "ros-"+workspace); err != nil {
			return err
		}
		return nil
	},
}
