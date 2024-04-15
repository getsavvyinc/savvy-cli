package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/idgen"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/shell/expansion"
	"github.com/spf13/cobra"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use: "send",
	// Send is meant for internal use only.
	Hidden: true,
	Short:  "Send data to the unix socket listening at SAVVY_SOCKET_PATH",
	Long:   `Send data to the unix socket listening at SAVVY_SOCKET_PATH`,
	Run: func(cmd *cobra.Command, args []string) {
		socketPath := os.Getenv("SAVVY_SOCKET_PATH")
		if socketPath == "" {
			err := fmt.Errorf("cannot record commands: SAVVY_SOCKET_PATH is not set")
			display.ErrorWithSupportCTA(err)
			return
		}
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			err = fmt.Errorf("failed to record command: %v", err)
			display.ErrorWithSupportCTA(err)
			return
		}
		defer conn.Close()
		message := strings.Join(args[:], " ") // ["echo hello world"] -> "echo hello world"
		if len(message) == 0 && (exitCode == 0 && stepID == "") {
			// nothing to do.
			return
		}

		message = expansion.IgnoreGrep(message)

		var quite bool
		if stepID == "" {
			stepID = idgen.New(idgen.CommandPrefix)
		} else {
			quite = true
		}

		data := server.RecordedData{
			Command:  message,
			StepID:   stepID,
			ExitCode: exitCode,
			Prompt:   prompt,
		}

		if err := json.NewEncoder(conn).Encode(data); err != nil {
			err = fmt.Errorf("failed to record command locally: %v", err)
			display.ErrorWithSupportCTA(err)
			return
		}

		if !quite {
			fmt.Print(stepID)
		}
	},
}

var stepID string
var exitCode int
var prompt string

func init() {
	rootCmd.AddCommand(sendCmd)
	// add flags to accept step_id as string and
	// exit_code as int from the command line
	sendCmd.Flags().StringVar(&stepID, "step-id", "", "Step ID")
	sendCmd.Flags().IntVar(&exitCode, "exit-code", 0, "Exit code")
	sendCmd.Flags().StringVar(&prompt, "prompt", "", "record shell prompt while command is executed")
}
