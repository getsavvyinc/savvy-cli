package display

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("2"))

func Success(text string) {
	fmt.Println(successStyle.Render(text))
}

func Successf(format string, args ...any) {
	text := fmt.Sprintf(format, args...)
	fmt.Println(successStyle.Render(text))
}
