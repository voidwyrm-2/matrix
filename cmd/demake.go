package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidwyrm-2/matrix/api/modpack"
)

var demakeCmd = &cobra.Command{
	Use:   "demake",
	Short: "Generates a Matrixfile from a matrix.toml",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		pack, err := modpack.FromToml("matrix.toml", false, false)
		if err != nil {
			return err
		}

		mods := []string{}

		for _, mod := range pack.Mods() {
			entry := ""
			p := mod.ToPublic()

			if p.Slug != "" {
				entry += p.Slug
			} else if p.Id != "" {
				entry += "id " + p.Id
			} else {
				continue
			}

			if p.ForceVersion != "" {
				entry += " " + p.ForceVersion
			}

			mods = append(mods, entry)
		}

		f, err := os.Create("matrixfile")
		if err != nil {
			return err
		}

		defer f.Close()

		_, err = f.Write([]byte(fmt.Sprintf("%s\n%s\n%s\n%s\n\n", pack.Name(), pack.Version(), pack.GameVersion(), pack.Modloader()) + strings.Join(mods, "\n")))
		return err
	},
}

func init() {
	rootCmd.AddCommand(demakeCmd)
}
