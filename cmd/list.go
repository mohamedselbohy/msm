package cmd

import (
	"fmt"
	"strconv"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

func colorStatus(status string, workspace string) string {
	switch status {
	case "active":
		return "\033[32m" + "(" + status + ")" + "\033[0m"
	case "inactive":
		return "\033[90m" + "(" + status + ")\033[0m" + "\033[90m: activate with (msm activate " + workspace + ")\033[0m"
	case "irresponsive":
		return "\033[31m" + "(" + status + ")\033[0m" + "\033[90m: Revive with (msm revive " + workspace + ")\033[0m"
	default:
		return status
	}
}

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lists open workspaces",
	Long:  "Lists open workspaces",
	Run: func(cmd *cobra.Command, args []string) {
		cli, ctx, err := docker.GetClient()
		if err != nil {
			fmt.Println("Error connecting to docker:", err)
			return
		}
		workspaces, err := docker.ListRunningContainers(cli, ctx)
		if err != nil {
			fmt.Println("Failed to list workspaces:", err)
			return
		}
		if len(workspaces) == 0 {
			fmt.Println("There are no active workspaces")
			return
		}
		for i, ws := range workspaces {
			fmt.Println(strconv.Itoa(i+1) + ": " + ws + "\t" + colorStatus(docker.IsActive(cli, ctx, "ros-"+ws), ws))
		}
	},
}
