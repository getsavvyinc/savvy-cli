package cmd

import (
	"fmt"
	"io"
	"log/slog"
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
		f, err := os.OpenFile("/Users/shantanu/.savvy_history", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		h := slog.NewTextHandler(f, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger := slog.New(h)
		logger.Debug("send called")
		defer f.Close()

		socketPath := os.Getenv("SAVVY_SOCKET_PATH")
		if socketPath == "" {
			socketPath = "/tmp/savvy-socket"
			// err := fmt.Errorf("cannot record commands: SAVVY_SOCKET_PATH is not set")
			// display.ErrorWithSupportCTA(err)
			// return
		}
		logger.Debug("dialing socket", "socketPath", socketPath)
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
		logger.Debug("writing to socket", "message", message)
		if _, err = io.WriteString(conn, message+"\n"); err != nil {
			err = fmt.Errorf("failed to record command locally: %v", err)
			display.ErrorWithSupportCTA(err)
			return
		}
		logger.Debug("finished writing to socket", "message", message)
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
