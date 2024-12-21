package model

type CodeInfo struct {
	Code     string            `json:"code"`
	Tags     map[string]string `json:"tags,omitempty"`
	FileData []byte            `json:"file_data,omitempty"`
	FileName string            `json:"file_name,omitempty"`
}
