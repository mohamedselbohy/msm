package cmd

import (
	"fmt"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var rvizCmd = &cobra.Command{
	Use:   "rviz",
	Short: "Launch rviz",
	Long:  "Launch rviz",
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
			fmt.Println("Error connecting to docker:", err)
		}
		if err = docker.ExecBackgroundCommand(cli, ctx, "ros-"+workspace, []string{"bash", "-c", "source /opt/ros/noetic/setup.bash && rviz"}); err != nil {
			fmt.Println("Failed connecting to container:", err)
		}
	},
}
