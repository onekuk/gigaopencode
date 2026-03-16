package main

import (
	"context"
	"errors"
	"io"

	"gitverse.ru/kmpavloff/openai-provider-gigachat/openai"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) CreateCompletion(ctx context.Context, req openai.CompletionRequest) (*openai.CompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (<-chan *openai.ChatCompletionStreamResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) ListModels(ctx context.Context) (*openai.ModelsResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) RetrieveModel(ctx context.Context, model string) (*openai.Model, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) DeleteModel(ctx context.Context, model string) error {
	return errors.New("not implemented")
}

func (p *MockProvider) ListFiles(ctx context.Context) (*openai.FilesResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) UploadFile(ctx context.Context, file io.Reader, filename, purpose string) (*openai.FileObject, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) RetrieveFile(ctx context.Context, fileID string) (*openai.FileObject, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) DeleteFile(ctx context.Context, fileID string) error {
	return errors.New("not implemented")
}

func (p *MockProvider) RetrieveFileContent(ctx context.Context, fileID string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateFineTuningJob(ctx context.Context, req openai.FineTuningJobRequest) (*openai.FineTuningJob, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) ListFineTuningJobs(ctx context.Context, after string, limit int) (*openai.FineTuningJobsResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) RetrieveFineTuningJob(ctx context.Context, jobID string) (*openai.FineTuningJob, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CancelFineTuningJob(ctx context.Context, jobID string) (*openai.FineTuningJob, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateImage(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateImageEdit(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateImageVariation(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateTranscription(ctx context.Context, req openai.AudioTranscriptionRequest) (*openai.AudioTranscriptionResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateTranslation(ctx context.Context, req openai.AudioTranslationRequest) (*openai.AudioTranslationResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateSpeech(ctx context.Context, model, voice, input string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (p *MockProvider) CreateModeration(ctx context.Context, req openai.ModerationRequest) (*openai.ModerationResponse, error) {
	return nil, errors.New("not implemented")
}