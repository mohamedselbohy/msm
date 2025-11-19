package cmd

import (
	"fmt"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "deactivates roscore for the workspace",
	Long:  "deactivates roscore for the workspace",
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
			`pkill -9 rosmaster`,
		})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Deactivated roscore for workspace:", name)
	},
}
