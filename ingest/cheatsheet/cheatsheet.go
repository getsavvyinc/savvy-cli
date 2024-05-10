package cheatsheet

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
