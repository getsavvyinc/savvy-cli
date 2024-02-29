package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/getsavvyinc/savvy-cli/display"
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
		if len(message) == 0 {
			// nothing to do.
			return
		}
		if _, err = io.WriteString(conn, message+"\n"); err != nil {
			err = fmt.Errorf("failed to record command locally: %v", err)
			display.ErrorWithSupportCTA(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
