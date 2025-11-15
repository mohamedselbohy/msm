package cmd

import "github.com/spf13/cobra"

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: "Manage ROS packages in the current workspace",
	Long:  "Manage ROS packages in the current workspace",
}

var (
	depsFile  string
	workspace string
)

func init() {
	pkgCreateCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace name")
	pkgCreateCmd.Flags().StringVarP(&depsFile, "dependencies", "d", "", "Dependencies file path")
	pkgCmd.AddCommand(pkgCreateCmd)
}
