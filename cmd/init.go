package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new ROS project",
	Long:  `Initialize a new ROS project Initializing a Docker Container with the same name given in the options`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if exists, err := docker.SearchRunningContainers("ros-" + name); err == nil {
			if exists {
				fmt.Println("Error: Workspace already exists")
				return
			}
		} else {
			fmt.Println("Error: Failed to search for containers")
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Failed to retrieve current working directory:", err)
			return
		}
		projectpath := filepath.Join(cwd, name)
		srcprojectpath := filepath.Join(projectpath, "src")
		_, err = os.Stat(srcprojectpath)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(srcprojectpath, 0o771); err != nil {
				fmt.Println("Failed to create directory:", err)
				return
			}
		}
		client, ctx, err := docker.GetClient()
		if err != nil {
			fmt.Println("Couldn't retrieve docker client:", err)
			return
		}
		docker.RunContainer(client, ctx, projectpath, name)
		output, err := docker.ExecCommand(client, ctx, "ros-"+name, []string{"bash", "-c", `echo -e "source /opt/ros/noetic/setup.bash\nsource /root/ros_ws/devel/setup.bash" >> ~/.bashrc`})
		if err != nil {
			fmt.Println("Execution Failed:", err)
			return
		}
		fmt.Println(string(output))
		output, err = docker.ExecCommand(client, ctx, "ros-"+name, []string{"bash", "-c", `source /opt/ros/noetic/setup.bash && cd /root/ros_ws && catkin_make`})
		if err != nil {
			fmt.Println("Execution Failed:", err)
			return
		}
		fmt.Println(string(output))
	},
}
