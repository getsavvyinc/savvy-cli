package theme

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func New() *huh.Theme {
	t := huh.ThemeDracula()

	light := catppuccin.Latte
	dark := catppuccin.Mocha
	var (
		subtext0 = lipgloss.AdaptiveColor{Light: light.Subtext0().Hex, Dark: dark.Subtext0().Hex}
		overlay1 = lipgloss.AdaptiveColor{Light: light.Overlay1().Hex, Dark: dark.Overlay1().Hex}
	)

	f := &t.Focused
	f.SelectedPrefix = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#02CF92", Dark: "#02A877"}).SetString("✓ ")
	f.UnselectedPrefix = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "", Dark: "243"}).SetString("• ")

	// light := catppuccin.Latte
	// dark := catppuccin.Mocha
	// green = lipgloss.AdaptiveColor{Light: light.Green().Hex, Dark: dark.Green().Hex}

	t.Help.Ellipsis.Foreground(subtext0)
	t.Help.ShortKey.Foreground(subtext0)
	t.Help.ShortDesc.Foreground(overlay1)
	t.Help.ShortSeparator.Foreground(subtext0)
	t.Help.FullKey.Foreground(subtext0)
	t.Help.FullDesc.Foreground(overlay1)
	t.Help.FullSeparator.Foreground(subtext0)

	return t
}
