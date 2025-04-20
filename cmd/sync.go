package cmd

import (
	"github.com/spf13/cobra"
	"github.com/voidwyrm-2/matrix/api/modpack"
)

var sync_ignoreNonempty, sync_ignoreExternals *bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Download all mods listed in the matrix.toml",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		pack, err := modpack.FromToml("matrix.toml", *sync_ignoreNonempty, *sync_ignoreExternals)
		if err != nil {
			return err
		}

		err = pack.Populate()
		if err != nil {
			return err
		}

		return pack.ToToml("matrix.toml")
	},
}

func init() {
	sync_ignoreNonempty = syncCmd.Flags().BoolP("empty", "e", false, "Only sync empty mods")
	sync_ignoreExternals = syncCmd.Flags().Bool("ext", false, "Don't attempt to download the external mods")

	rootCmd.AddCommand(syncCmd)
}
