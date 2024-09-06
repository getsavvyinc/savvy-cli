package setup

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed savvy.fish
var fishSetupFiles embed.FS

const fishSetupScriptName = "savvy.fish"

// fishCmd represents the fish command
var FishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Output shell setup for fish",
	Long:  `Output shell setup for bash`,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := fishSetupFiles.ReadFile(fishSetupScriptName)
		if err != nil {
			return err
		}
		fmt.Println(string(content))
		return nil
	},
}
