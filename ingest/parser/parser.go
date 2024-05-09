package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Parser interface {
	Parse(path string) (*CheatSheet, error)
	Provider() CheatSheetProvider
}

type CheatSheet struct {
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	Examples    []*Example `json:"examples,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
}

type Example struct {
	Explanation string `json:"explanation,omitempty"`
	Command     string `json:"command,omitempty"`
}

type CheatSheetProvider string

const (
	TLDR CheatSheetProvider = "tldr"
)

func New(provider CheatSheetProvider) Parser {
	switch provider {
	case TLDR:
		return &tldr{
			provider: provider,
		}
	default:
		return nil
	}
}

type tldr struct {
	provider CheatSheetProvider
}

var _ Parser = &tldr{}

var ErrRequiredMdFile = fmt.Errorf("required markdown file")

func (t *tldr) Parse(path string) (*CheatSheet, error) {
	if ext := filepath.Ext(path); ext != ".md" && ext != ".markdown" {
		return nil, fmt.Errorf("%w: invalid file extension: %s", ErrRequiredMdFile, ext)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	cs := &CheatSheet{}
	var description []string
	var explanations []string
	var commands []string

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "#"):
			if cs.Title != "" {
				// We've already set the title
				continue
			}
			cs.Title = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		case strings.HasPrefix(line, ">"):
			description = append(description, strings.TrimSpace(strings.TrimPrefix(line, ">")))
		case strings.HasPrefix(line, "-"):
			explanations = append(explanations, strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "-")), ":"))
		case strings.HasPrefix(line, "`"):
			commands = append(commands, strings.TrimSpace(strings.Trim(line, "`")))
		default:
			// empty lines, or lines that don't match the above
			continue
		}
	}

	cs.Description = strings.Join(description, " ")
	cs.Examples = zip(explanations, commands)
	return cs, nil
}

func (t *tldr) Provider() CheatSheetProvider {
	return t.provider
}

func zip(explanations, commands []string) []*Example {
	examples := make([]*Example, 0, len(explanations))
	for i := range explanations {
		examples = append(examples, &Example{
			Explanation: explanations[i],
			Command:     commands[i],
		})
	}
	return examples
}
