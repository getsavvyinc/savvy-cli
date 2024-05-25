package cmd

import (
	"errors"
	"fmt"
	"os"

	huhSpinner "github.com/charmbracelet/huh/spinner"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/spf13/cobra"
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write writes a file to your file system",
	Example: `
  savvy write --step-id=stp-id
  `,
	Long: `
  Write writes a file to your file system.

  Write is used while running a runbook to write important data to your filesystem.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cl, err := client.New()
		if err != nil && errors.Is(err, client.ErrInvalidClient) {
			display.Error(errors.New("You must be logged in to use savvy write. Please run `savvy login`"))
			os.Exit(1)
		}

		var stepContent *client.StepContent

		if err := huhSpinner.New().Title("Fetching recorded data...").Action(func() {
			var err error
			stepContent, err = cl.StepContentByStepID(cmd.Context(), writeStepID)

			if err != nil {
				display.FatalErrWithSupportCTA(err)
				return
			}
		}).Run(); err != nil {
			display.FatalErrWithSupportCTA(err)
			return
		}

		fmt.Println("StepContent:\n", string(stepContent.Content))
	},
}

var writeStepID string

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.Flags().StringVar(&sendStepID, "step-id", "", "The step id linked to the file")
}
