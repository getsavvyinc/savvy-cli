package llm

import (
	"io/fs"
)

type RecordedCommand struct {
	Command  string    `json:"command"`
	Prompt   string    `json:"prompt,omitempty"`
	FileInfo *FileInfo `json:"file_info,omitempty"`
}

type FileInfo struct {
	Mode    fs.FileMode `json:"mode,omitempty"`
	Content []byte      `json:"content,omitempty"`
	Path    string      `json:"path,omitempty"`
}
