package cmd

import (
	"fmt"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "activates roscore for the workspace",
	Long:  "activates roscore for the workspace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if exists, err := docker.SearchRunningContainers("ros-" + name); err == nil {
			if !exists {
				fmt.Println("Error: Workspace does not exist.")
				return
			}
		} else {
			fmt.Println("Error: Failed to search for containers")
			return
		}
		cli, ctx, err := docker.GetClient()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		err = docker.ExecBackgroundCommand(cli, ctx, "ros-"+name, []string{
			"bash", "-c",
			`source /opt/ros/noetic/setup.bash && roscore`,
		})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Activated roscore for workspace:", name)
	},
}
