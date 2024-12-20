package model

type QuestionInfo struct {
	Question          string            `json:"question"`
	Tags              map[string]string `json:"tags,omitempty"`
	FileData          []byte            `json:"file_data,omitempty"`
	FileName          string            `json:"file_name,omitempty"`
	PreviousQuestions []string          `json:"previous_questions,omitempty"`
	PreviousCommands  []string          `json:"previous_commands,omitempty"`
}
