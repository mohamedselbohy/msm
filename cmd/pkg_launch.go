package cmd

import (
	"fmt"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var pkgLaunchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a package from a launch file",
	Long:  "Run a package from a launch file",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if workspace == "" {
			return fmt.Errorf("error: You must select a workspace")
		}
		if exists, err := docker.SearchRunningContainers("ros-" + workspace); err == nil {
			if !exists {
				return fmt.Errorf("error: Workspace does not exist")
			}
		} else {
			return err
		}
		cli, ctx, err := docker.GetClient()
		if err != nil {
			return err
		}
		var executable string
		pkg := args[0]
		if len(args) > 1 {
			executable = args[1]
		}
		err = docker.EngageCommand(cli, ctx, "ros-"+workspace, []string{
			"bash", "-c",
			`source /opt/ros/noetic/setup.bash && source /root/ros_ws/devel/setup.bash && roslaunch ` + pkg + ` ` + executable,
		})
		if err != nil {
			return err
		}
		return nil
	},
}
