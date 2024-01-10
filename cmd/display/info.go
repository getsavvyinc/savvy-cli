package display

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var infoStyle = lipgloss.NewStyle().
	Bold(false).
	PaddingTop(1).
	Foreground(lipgloss.Color("21"))

func Info(text string) {
	fmt.Println(infoStyle.Render(text))
}
