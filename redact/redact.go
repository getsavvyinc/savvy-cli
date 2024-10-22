package redact

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/slice"
	"github.com/getsavvyinc/savvy-cli/theme"
)

// Commands allows users to redact one or more commands.
// Commands returns a new slice of commands with the sensitive data redacted or removed.
//
// NOTE: It is possible for users to completely remove some commands.
func Commands(cmds []*server.RecordedCommand) ([]*server.RecordedCommand, error) {
	var fs []huh.Field
	description := "Replace sensitive data with <placeholders>. To remove a command, simply delete the text."
	note := huh.NewNote().Title("Redact Secrets and PII").Description(description)
	fs = append(fs, note)

	for i, cmd := range cmds {
		fs = append(fs, RedactCommand(cmd.Command, strconv.Itoa(i)))
	}

	customTheme := theme.New()

	group := huh.NewGroup(fs...).Title("Redact Commands").WithTheme(customTheme)

	if err := huh.NewForm(group).WithTheme(customTheme).Run(); err != nil {
		err := fmt.Errorf("failed to run redaction form: %w", err)
		return nil, err
	}

	for _, f := range fs {
		in, ok := f.(*huh.Input)
		if !ok {
			continue
		}
		strVal, ok := in.GetValue().(string)
		if !ok {
			continue
		}

		idx, err := strconv.Atoi(in.GetKey())
		if err != nil {
			continue
		}
		cmds[idx].Command = strVal
	}

	redacted := slice.Filter(cmds, func(cmd *server.RecordedCommand) bool {
		return cmd.Command != "" || cmd.FileInfo != nil
	})
	return redacted, nil
}

func RedactCommand(cmd string, key string) huh.Field {
	return huh.NewInput().Value(&cmd).Key(key).WithTheme(theme.New())
}
