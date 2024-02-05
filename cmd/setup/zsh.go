package setup

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed savvy.zsh
var zshSetupScript embed.FS

const zshSetupScriptName = "savvy.zsh"

// initCmd represents the init command
var ZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Output shell setup for zsh",
	Long:  `Output shell setup for zsh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := zshSetupScript.ReadFile(zshSetupScriptName)
		if err != nil {
			return err
		}
		fmt.Println(string(content))
		return nil
	},
}
