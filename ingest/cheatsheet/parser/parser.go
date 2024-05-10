package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet"
)

type Parser interface {
	Parse(path string) (*cheatsheet.CheatSheet, error)
	Provider() cheatsheet.Provider
}

func New(provider cheatsheet.Provider) Parser {
	switch provider {
	case cheatsheet.TLDR:
		return &tldr{
			provider: provider,
		}
	default:
		return nil
	}
}

type tldr struct {
	provider cheatsheet.Provider
}

var _ Parser = &tldr{}

var ErrRequiredMdFile = fmt.Errorf("required markdown file")

func (t *tldr) Parse(path string) (*cheatsheet.CheatSheet, error) {
	if ext := filepath.Ext(path); ext != ".md" && ext != ".markdown" {
		return nil, fmt.Errorf("%w: invalid file extension: %s", ErrRequiredMdFile, ext)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	cs := &cheatsheet.CheatSheet{}

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

	dir, _ := filepath.Split(path)
	// If the directory is not empty and not common, add it to the tags.
	if tag := filepath.Base(dir); tag != "" && tag != "common" {
		cs.Tags = append(cs.Tags, tag)
	}

	cs.Description = strings.Join(description, " ")
	cs.Examples = zip(explanations, commands)
	return cs, nil
}

func (t *tldr) Provider() cheatsheet.Provider {
	return t.provider
}

func zip(explanations, commands []string) []*cheatsheet.Example {
	examples := make([]*cheatsheet.Example, 0, len(explanations))
	for i := range explanations {
		examples = append(examples, &cheatsheet.Example{
			Explanation: explanations[i],
			Command:     commands[i],
		})
	}
	return examples
}
