package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/voidwyrm-2/matrix/api/modpack"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all non-external mods currently downloaded",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		pack, err := modpack.FromToml("matrix.toml", false, false)
		if err != nil {
			return err
		}

		for _, mod := range pack.Mods() {
			if !mod.IsEmpty() {
				fmt.Println(mod.Name())
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
