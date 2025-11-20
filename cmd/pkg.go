package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mohamedselbohy/msm/internal/docker"
	"github.com/spf13/cobra"
)

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Manage ROS packages in the current workspace",
	Long:  "Manage ROS packages in the current workspace",
}

var (
	depsFile  string
	workspace string
)

func CompletePackagesAndExecutables(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, ctx, err := docker.GetClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	output, err := docker.ExecCommand(cli, ctx, "ros-"+workspace, []string{
		"bash", "-c",
		`source /opt/ros/noetic/setup.bash && source /root/ros_ws/devel/setup.bash && rospack list`,
	})
	if err != nil {
		fmt.Println("error")
		return nil, cobra.ShellCompDirectiveError
	}
	lines := strings.Split(output, "\n")
	packages := make(map[string]string)
	var packs []string
	for _, line := range lines {
		words := strings.Split(line, " ")
		packs = append(packs, words[0])
		if len(words) > 1 {
			packages[words[0]] = words[1]
		}
	}
	var completions []string
	switch len(args) {
	case 0:
		for _, pack := range packs {
			if strings.HasPrefix(pack, toComplete) {
				completions = append(completions, pack)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault
	case 1:
		output, err = docker.ExecCommand(cli, ctx, "ros-"+workspace, []string{
			"bash", "-c",
			`find ` + packages[args[0]] + ` /opt/ros/noetic/lib/` + args[0] + `  -type f -executable 2>/dev/null`,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		lines = strings.Split(output, "\n")
		var executables []string
		for _, line := range lines[:len(lines)-1] {
			executables = append(executables, filepath.Base(line))
		}
		for _, ex := range executables {
			if strings.HasPrefix(ex, toComplete) {
				completions = append(completions, ex)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func CompletePackagesAndLaunches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, ctx, err := docker.GetClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	output, err := docker.ExecCommand(cli, ctx, "ros-"+workspace, []string{
		"bash", "-c",
		`source /opt/ros/noetic/setup.bash && source /root/ros_ws/devel/setup.bash && rospack list`,
	})
	if err != nil {
		fmt.Println("error")
		return nil, cobra.ShellCompDirectiveError
	}
	lines := strings.Split(output, "\n")
	packages := make(map[string]string)
	var packs []string
	for _, line := range lines {
		words := strings.Split(line, " ")
		packs = append(packs, words[0])
		if len(words) > 1 {
			packages[words[0]] = words[1]
		}
	}
	var completions []string
	switch len(args) {
	case 0:
		for _, pack := range packs {
			if strings.HasPrefix(pack, toComplete) {
				completions = append(completions, pack)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault
	case 1:
		output, err = docker.ExecCommand(cli, ctx, "ros-"+workspace, []string{
			"bash", "-c",
			`find ` + packages[args[0]] + ` /opt/ros/noetic/share/` + args[0] + `  -type f -name "*.launch*" 2>/dev/null`,
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		lines = strings.Split(output, "\n")
		var executables []string
		for _, line := range lines[:len(lines)-1] {
			executables = append(executables, filepath.Base(line))
		}
		for _, ex := range executables {
			if strings.HasPrefix(ex, toComplete) {
				completions = append(completions, ex)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func init() {
	pkgCreateCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace name")
	pkgCreateCmd.RegisterFlagCompletionFunc("workspace", CompleteWorkspaces)
	pkgRunCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace name")
	pkgRunCmd.RegisterFlagCompletionFunc("workspace", CompleteWorkspaces)
	pkgLaunchCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace name")
	pkgLaunchCmd.RegisterFlagCompletionFunc("workspace", CompleteWorkspaces)
	pkgRunCmd.ValidArgsFunction = CompletePackagesAndExecutables
	pkgLaunchCmd.ValidArgsFunction = CompletePackagesAndLaunches
	pkgCreateCmd.Flags().StringVarP(&depsFile, "dependencies", "d", "", "Dependencies file path")
	pkgCmd.AddCommand(pkgCreateCmd)
	pkgCmd.AddCommand(pkgRunCmd)
	pkgCmd.AddCommand(pkgLaunchCmd)
}
