package savvy_errors

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("9"))

// Display prints the error and any additional messages to the terminal
func Display(err error, msgs ...string) {
	// be defensive
	if err == nil {
		return
	}

	errMsg := err.Error()
	if errMsg == "" {
		return
	}

	fmt.Println(style.Render(err.Error()))
	for _, msg := range msgs {
		fmt.Println(style.Render(msg))
	}
}

const supportCTA = `Stuck? We're here to make things easiser for you. Just email us at support@getsavvy.so or join our friendly Discord community (https://getsavvy.so/discord) for a chat.`

func DisplayWithSupportCTA(err error) {
	Display(err, supportCTA)
}
