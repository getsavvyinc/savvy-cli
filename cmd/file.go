package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Record contents of a file given its path",
	Long:  `Record contents and filename of a file given its path.`,
	Run: func(cmd *cobra.Command, args []string) {
		_, err := client.New()
		if err != nil && errors.Is(err, client.ErrInvalidClient) {
			display.Error(errors.New("You must be logged in to record a runbook. Please run `savvy login`"))
			os.Exit(1)
		}

		if len(args) == 0 {
			display.ErrorMsg("no file path provided")
			return
		}

		if len(args) > 1 {
			display.ErrorMsg("file only accepts one filepath at a time")
			return
		}

		filePath := args[0]
		fi, err := os.Stat(filePath)
		if err != nil {
			display.Error(err)
			return
		}

		if fi.IsDir() {
			display.ErrorMsg("file path provided is a directory")
			return
		}

		if fi.Size() == 0 {
			display.ErrorMsg("file provided is empty")
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			display.Error(err)
			return
		}
		defer file.Close()

		cl, err := server.NewDefaultClient(context.Background())
		if err != nil {
			display.Error(err)
			return
		}

		command, err := recordFile(file, fi)
		if err != nil {
			display.Error(err)
			return
		}

		if err := cl.Send(command); err != nil {
			display.ErrorWithSupportCTA(err)
			return
		}
	},
}

func init() {
	recordCmd.AddCommand(fileCmd)
}

func recordFile(file *os.File, fi os.FileInfo) (string, error) {
	bs, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	fileMetadata := fmt.Sprintf("Filename: %s\nPermission: %s", fi.Name(), fi.Mode().String())

	command := fmt.Sprintf("%s\n%s", fileMetadata, string(bs))

	return command, nil
}
