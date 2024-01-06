package savvy_errors

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	PaddingTop(1).
	Foreground(lipgloss.Color("9"))

func Display(err error) {
	fmt.Println(style.Render(err.Error()))
}
