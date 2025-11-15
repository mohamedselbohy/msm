package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var pkgCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new package in a specific workspace",
	Long:  "Create new package in a specific workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if workspace == "" {
			return fmt.Errorf("you must choose a workspace")
		}
		if exists, err := docker.SearchRunningContainers("ros-" + workspace); err == nil {
			if !exists {
				return fmt.Errorf("error: Workspace is not active")
			}
		} else {
			return fmt.Errorf("error: Failed to search for containers")
		}
		var deps []string
		if len(args) > 1 {
			deps = args[1:]
		} else if depsFile != "" {
			content, err := os.ReadFile(depsFile)
			if err != nil {
				return fmt.Errorf("failed to read dependencies file: %w", err)
			}
			deps = strings.Split(string(content[:len(content)-1]), "\n")
		}
		depsLine := strings.Join(deps, " ")
		cli, ctx, err := docker.GetClient()
		if err != nil {
			return err
		}
		output, err := docker.ExecCommand(cli, ctx, "ros-"+workspace, []string{"bash", "-c", `source /opt/ros/noetic/setup.bash && cd /root/ros_ws/src && catkin_create_pkg ` + name + " " + depsLine})
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	},
}
