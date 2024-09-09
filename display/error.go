package display

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("9"))

// Error prints the error and any additional messages to the terminal
func Error(err error, msgs ...string) {
	// be defensive
	if err == nil {
		return
	}

	errMsg := err.Error()
	if errMsg == "" {
		return
	}

	ErrorMsg(err.Error())
	if len(msgs) > 0 {
		ErrorMsg(msgs...)
	}
}

func ErrorMsg(msgs ...string) {
	for _, msg := range msgs {
		fmt.Println(style.Render(msg))
	}
}

func FatalErr(err error, msgs ...string) {
	Error(err, msgs...)
	os.Exit(1)
}

func FatalErrWithSupportCTA(err error) {
	Error(err, supportCTA)
	os.Exit(1)
}

const supportCTA = `Stuck? We're here to make things easier for you. Just email us at support@getsavvy.so or join our friendly Discord community (https://getsavvy.so/discord) for a chat.`

func ErrorWithSupportCTA(err error) {
	Error(err, supportCTA)
}
