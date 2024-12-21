package llm

import (
	"errors"

	"github.com/sashabaranov/go-openai"
)

type ResponseStreamer interface {
	Recv() ([]byte, error)
	Close() error
}

type StreamData struct {
	Data string `json:"data"`
}

type streamer struct {
	stream *openai.ChatCompletionStream
}

// NewStreamer creates a new streamer.
func NewStreamer(stream *openai.ChatCompletionStream) ResponseStreamer {
	return &streamer{stream: stream}
}

// Recv reads the next response from the stream.
// Recv blocks until it receives a response or an error occurs.
// Recv returns io.EOF when the stream has been closed.
func (s *streamer) Recv() ([]byte, error) {
	completion, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}

	if len(completion.Choices) == 0 {
		return nil, errors.New("no completions returned")
	}

	data := completion.Choices[0].Delta.Content

	return []byte(data), nil
}

// Close closes the stream and releases any resources associated with it.
// Close should be called when the caller is done with the stream.
func (s *streamer) Close() error {
	if s.stream == nil {
		return nil
	}
	return s.stream.Close()
}
