package cleanup

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/getsavvyinc/savvy-cli/server/mode"
)

func GetPermission(m mode.Mode) (bool, error) {
	var confirmation bool
	confirmCleanup := huh.NewConfirm().
		Title(fmt.Sprintf("Multiple %s sessions detected", m)).
		Affirmative("Continue here and kill other sessions").
		Negative("Quit this session").
		Value(&confirmation)
	if err := huh.NewForm(huh.NewGroup(confirmCleanup)).Run(); err != nil {
		return false, err
	}
	return confirmation, nil
}
