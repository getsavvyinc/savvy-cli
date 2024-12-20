package llm

type Runbook struct {
	Title string
	Steps []RunbookStep
}

type StepTypeEnum string

const (
	StepTypeCode StepTypeEnum = "code"
	StepTypeFile StepTypeEnum = "file"
)

type RunbookStep struct {
	Type        StepTypeEnum `json:"type"`
	Description string       `json:"description"`
	Command     string       `json:"command"`
	CommandID   string       `json:"command_id"`
}
