package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "matrix",
	Short: "Matrix is a Minecraft mod manager for Modrinth",
	Long:  ``,
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func init() {
}
