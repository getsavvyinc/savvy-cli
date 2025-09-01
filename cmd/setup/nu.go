package setup

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed savvy.nu
var nuSetupScript embed.FS

const nuSetupScriptName = "savvy.nu"

// NuCmd represents the nu command
var NuCmd = &cobra.Command{
	Use:   "nu",
	Short: "Output shell setup for nushell",
	Long:  `Output shell setup for nushell`,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := nuSetupScript.ReadFile(nuSetupScriptName)
		if err != nil {
			return err
		}
		fmt.Println(string(content))
		return nil
	},
}
