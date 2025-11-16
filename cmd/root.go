package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "msm",
	Short: "A ROS interface, to deal with docker-based ros environments",
	Long:  `A ROS interface, to deal with docker-based ros environments`,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(pkgCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
