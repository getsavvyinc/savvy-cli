package llm

type Runbook struct {
	Title string
	Steps []Step
}

type StepTypeEnum string

const (
	StepTypeCode StepTypeEnum = "code"
	StepTypeFile StepTypeEnum = "file"
)

type Step struct {
	Type        StepTypeEnum `json:"type"`
	Description string       `json:"description"`
	Command     string       `json:"command"`
}
