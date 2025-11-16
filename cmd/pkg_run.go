package cmd

import (
	"fmt"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var pkgRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a package with executable",
	Long:  "Run a package with executable",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if workspace == "" {
			return fmt.Errorf("error: Wou must select a workspace")
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
		err = docker.ExecBackgroundCommand(cli, ctx, "ros-"+workspace, []string{
			"bash", "-c",
			`source /opt/ros/noetic/setup.bash && roscore`,
		})
		if err != nil {
			return err
		}
		var executable string
		pkg := args[0]
		if len(args) > 1 {
			executable = args[1]
		} else {
			executable = "main.py"
		}
		output, err := docker.ExecCommand(cli, ctx, "ros-"+workspace, []string{
			"bash", "-c",
			`source /opt/ros/noetic/setup.bash && source /root/ros_ws/devel/setup.bash && rosrun ` + pkg + ` ` + executable,
		})
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		_, err = docker.ExecCommand(cli, ctx, "ros-"+workspace, []string{
			"bash", "-c",
			`source /opt/ros/noetic/setup.bash && pkill -9 rosmaster && pkill -9 rosout`,
		})
		if err != nil {
			return err
		}
		return nil
	},
}
