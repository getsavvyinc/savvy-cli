package display

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var infoStyle = lipgloss.NewStyle().
	Bold(false).
	PaddingTop(1).
	PaddingBottom(1)

func Info(text string) {
	fmt.Println(infoStyle.Render(text))
}

func Infof(format string, a ...interface{}) {
	Info(fmt.Sprintf(format, a...))
}
