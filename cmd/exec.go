package cmd

import (
	"fmt"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Exec into workspace ROS",
	Long:  "Exec into workspace ROS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspace := args[0]
		if exists, err := docker.SearchRunningContainers("ros-" + workspace); err == nil {
			if !exists {
				fmt.Println("Error: Workspace is not active:", err)
			}
		} else {
			fmt.Println("Error: Failed to search for containers:", err)
		}

		cli, ctx, err := docker.GetClient()
		if err != nil {
			fmt.Println("Error connecting to workspace:", err)
		}
		if err = docker.ExecIntoContainer(cli, ctx, "ros-"+workspace); err != nil {
			fmt.Println("Failed connecting to container:", err)
		}
	},
}
