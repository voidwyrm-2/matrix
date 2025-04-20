package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current Matrix version",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Matrix mod manager version", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
