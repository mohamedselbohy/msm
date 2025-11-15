package cmd

import (
	"fmt"
	"os"
	"strings"

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
		fmt.Println("Workspace:", workspace)
		fmt.Println("Creating package:", name)
		fmt.Println("Dependencies:", deps)
		return nil
	},
}
