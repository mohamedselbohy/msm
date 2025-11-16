package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "msm",
	Short: "A ROS interface, to deal with docker-based ros environments",
	Long:  `A ROS interface, to deal with docker-based ros environments`,
}

func CompleteWorkspaces(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, ctx, err := docker.GetClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	workspaces, err := docker.ListRunningContainers(cli, ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var completions []string
	for _, ws := range workspaces {
		if strings.HasPrefix(ws, toComplete) {
			completions = append(completions, ws)
		}
	}
	return completions, cobra.ShellCompDirectiveDefault
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(pkgCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	execCmd.ValidArgsFunction = CompleteWorkspaces
	removeCmd.ValidArgsFunction = CompleteWorkspaces
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
