package cmd

import (
	"fmt"
	"os"

	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/login"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to savvy",
	Long:  `Login allows users to use Google SSO to login to savvy.`,
	Run:   runLoginCmd,
}

func runLoginCmd(cmd *cobra.Command, args []string) {
	force, err := cmd.Flags().GetBool(forceLoginFlag)
	if err != nil {
		display.ErrorWithSupportCTA(fmt.Errorf("error parsing flags: %w", err))
		os.Exit(1)
	}

	if err := login.Verify(); err == nil && !force {
		display.Info("You are already logged in!")
		display.Info("Run `savvy login --force` to get a new token")
		return
	}

	login.Run()
}

const forceLoginFlag = "force"
const forceLoginFlagShort = "f"

func init() {
	loginCmd.Flags().BoolP(forceLoginFlag, forceLoginFlagShort, false, "Force new login flow")
	rootCmd.AddCommand(loginCmd)
}
