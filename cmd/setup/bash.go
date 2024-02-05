package setup

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed bash-preexec.sh
var bashSetupScript embed.FS

const bashSetupScriptName = "bash-preexec.sh"

// initCmd represents the init command
var BashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Output shell setup for bash",
	Long:  `Output shell setup for bash`,
	RunE:  runCmd,
}

func runCmd(cmd *cobra.Command, args []string) error {
	content, err := bashSetupScript.ReadFile(bashSetupScriptName)
	if err != nil {
		return err
	}
	fmt.Println(string(content))
	return nil
}

var DashCmd = &cobra.Command{
	Use:   "dash",
	Short: "Output shell setup for dash",
	Long:  `Output shell setup for dash`,
	RunE:  runCmd,
}
