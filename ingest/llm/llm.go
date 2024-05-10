package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type Client interface {
	CreateEmbeddings(ctx context.Context, input string) ([]float32, error)
}

const (
	embeddingModelName  = "text-embedding-3-small"
	embeddingDimensions = 512
)

type openAI struct {
	client *openai.Client
}

func NewOpenAIClient(authToken string) Client {
	client := openai.NewClient(authToken)
	return &openAI{
		client: client,
	}
}

func (o *openAI) CreateEmbeddings(ctx context.Context, input string) ([]float32, error) {
	resp, err := o.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model:          embeddingModelName,
		Input:          input,
		Dimensions:     512,
		EncodingFormat: openai.EmbeddingEncodingFormatFloat,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) != 1 {
		return nil, fmt.Errorf("expected 1 embedding, got %d", len(resp.Data))
	}
	return resp.Data[0].Embedding, nil
}
