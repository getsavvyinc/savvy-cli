package history

import (
	"strings"
	"unicode"
)

type Command struct {
	command string
}

func (c *Command) GetCommand() string {
	return c.command
}

func (c *Command) String() string {
	return c.command
}

func NewCommand(command string) *Command {
	return &Command{command: command}
}

type Parser interface {
	Parse(history []string) []*Command
}

func NewParser() Parser {

	return &parser{}
}

type parser struct{}

var _ Parser = &parser{}

// ParseHistory parses the history and returns a slice of Command objects.
// ParseHistory's implementaiton supports the following history fomat:
// |\s+<index>\s+<command>
func (p *parser) ParseHistory(history []string) []*Command {
	var commands []*Command
	if len(history) == 0 {
		return nil
	}

	for _, h := range history {
		noSurroundingSpace := strings.TrimSpace(h)
		parts := strings.FieldsFunc(noSurroundingSpace, func(r rune) bool {
			return unicode.IsSpace(r)
		})
		if len(parts) == 0 {
			continue
		} else if len(parts) == 1 {
			commands = append(commands, &Command{command: parts[0]})
		} else {
			// If the history is not in the expected format, we will just add the command as is.
			commands = append(commands, &Command{command: strings.Join(parts[1:], " ")})
		}
	}
	return commands
}
