package cheatsheet

import "strings"

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

type Provider string

const (
	TLDR Provider = "tldr"
)

func (cs *CheatSheet) CommonEmbeddingPrefix() string {
	tags := strings.Join(cs.Tags, ",")
	return strings.Join([]string{cs.Title, cs.Description, tags}, " ")
}
