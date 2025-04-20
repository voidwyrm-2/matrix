package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voidwyrm-2/matrix/api/modpack"
)

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Generate a matrix.toml from a Matrixfile",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return modpack.FromMatrixfile("matrix.toml")
	},
}

func init() {
	rootCmd.AddCommand(makeCmd)
}
