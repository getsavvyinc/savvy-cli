package display

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var infoStyle = lipgloss.NewStyle().
	Bold(false).
	PaddingTop(1).
	PaddingBottom(1).
	Foreground(lipgloss.AdaptiveColor{
		Light: "21",
		Dark:  "33",
	})

func Info(text string) {
	fmt.Println(infoStyle.Render(text))
}
