package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
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

		// Write to the file system at the specified place.
		// Allow the user to overwrite the name/path of the file
		filename := stepContent.Name

		dirPath := stepContent.DirPath
		if len(stepContent.DirPath) == 0 {
			dirPath = "."
		}

		maybeEditFileName := huh.NewInput().
			Title("Confirm or Edit File name").
			Description("Press enter to confirm or edit the file name").
			Prompt("> ").
			Value(&filename)

		maybeEditDirPath := huh.NewInput().
			Title("Confrim or Edit Directory Path").
			Prompt("> ").
			Value(&dirPath)

		form := huh.NewForm(huh.NewGroup(maybeEditFileName, maybeEditDirPath))
		if err := form.Run(); err != nil {
			display.FatalErrWithSupportCTA(err)
			return
		}

		buf := bytes.NewBuffer(stepContent.Content)
		writeFile(filepath.Join(dirPath, filename), buf)

	},
}

var writeStepID string

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.Flags().StringVar(&writeStepID, "step-id", "", "The step id linked to the file")
}

func writeFile(path string, buf *bytes.Buffer) {
	// TODO: create dir if required
	// TODO: use the same mode as the original file
	// TODO: confrim with the user before overwriting
	f, err := os.Create(path)
	if err != nil {
		display.Error(err)
		return
	}
	defer f.Close()

	_, err = buf.WriteTo(f)
	if err != nil {
		display.Error(err)
		return
	}

	display.Info("File written to " + path)
}
