package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"gitverse.ru/kmpavloff/openai-provider-gigachat/openai"
)

type Provider interface {
	CompletionHandler
	ChatCompletionHandler
	EmbeddingHandler
	ModelHandler
	FileHandler
	FineTuneHandler
	ImageHandler
	AudioHandler
	ModerationHandler
}

type CompletionHandler interface {
	CreateCompletion(ctx context.Context, req openai.CompletionRequest) (*openai.CompletionResponse, error)
}

type ChatCompletionHandler interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
	CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (<-chan *openai.ChatCompletionStreamResponse, error)
}

type EmbeddingHandler interface {
	CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error)
}

type ModelHandler interface {
	ListModels(ctx context.Context) (*openai.ModelsResponse, error)
	RetrieveModel(ctx context.Context, model string) (*openai.Model, error)
	DeleteModel(ctx context.Context, model string) error
}

type FileHandler interface {
	ListFiles(ctx context.Context) (*openai.FilesResponse, error)
	UploadFile(ctx context.Context, file io.Reader, filename, purpose string) (*openai.FileObject, error)
	RetrieveFile(ctx context.Context, fileID string) (*openai.FileObject, error)
	DeleteFile(ctx context.Context, fileID string) error
	RetrieveFileContent(ctx context.Context, fileID string) (io.ReadCloser, error)
}

type FineTuneHandler interface {
	CreateFineTuningJob(ctx context.Context, req openai.FineTuningJobRequest) (*openai.FineTuningJob, error)
	ListFineTuningJobs(ctx context.Context, after string, limit int) (*openai.FineTuningJobsResponse, error)
	RetrieveFineTuningJob(ctx context.Context, jobID string) (*openai.FineTuningJob, error)
	CancelFineTuningJob(ctx context.Context, jobID string) (*openai.FineTuningJob, error)
}

type ImageHandler interface {
	CreateImage(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error)
	CreateImageEdit(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error)
	CreateImageVariation(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error)
}

type AudioHandler interface {
	CreateTranscription(ctx context.Context, req openai.AudioTranscriptionRequest) (*openai.AudioTranscriptionResponse, error)
	CreateTranslation(ctx context.Context, req openai.AudioTranslationRequest) (*openai.AudioTranslationResponse, error)
	CreateSpeech(ctx context.Context, model, voice, input string) (io.ReadCloser, error)
}

type ModerationHandler interface {
	CreateModeration(ctx context.Context, req openai.ModerationRequest) (*openai.ModerationResponse, error)
}

type HTTPHandlers struct {
	provider Provider
}

func NewHTTPHandlers(provider Provider) *HTTPHandlers {
	return &HTTPHandlers{
		provider: provider,
	}
}

func (h *HTTPHandlers) HandleCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req openai.CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	result, err := h.provider.CreateCompletion(r.Context(), req)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *HTTPHandlers) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Stream {
		ch, err := h.provider.CreateChatCompletionStream(r.Context(), req)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "internal_error", err.Error())
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		for data := range ch {
			jsonData, _ := json.Marshal(data)
			w.Write([]byte("data: "))
			w.Write(jsonData)
			w.Write([]byte("\n\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		w.Write([]byte("data: [DONE]\n\n"))
		return
	}

	result, err := h.provider.CreateChatCompletion(r.Context(), req)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *HTTPHandlers) HandleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	var req openai.EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	result, err := h.provider.CreateEmbeddings(r.Context(), req)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *HTTPHandlers) HandleListModels(w http.ResponseWriter, r *http.Request) {
	result, err := h.provider.ListModels(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *HTTPHandlers) HandleRetrieveModel(w http.ResponseWriter, r *http.Request) {
	modelID := extractIDFromPath(r.URL.Path, "/v1/models/")
	result, err := h.provider.RetrieveModel(r.Context(), modelID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *HTTPHandlers) HandleDeleteModel(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleRetrieveFile(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleRetrieveFileContent(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateFineTuningJob(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleListFineTuningJobs(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleRetrieveFineTuningJob(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCancelFineTuningJob(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateImage(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateImageEdit(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateImageVariation(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateTranscription(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateTranslation(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateSpeech(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func (h *HTTPHandlers) HandleCreateModeration(w http.ResponseWriter, r *http.Request) {
	// Implementation placeholder
	writeNotImplemented(w)
}

func writeNotImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": {"message": "Not implemented", "type": "not_implemented_error"}}`))
}
