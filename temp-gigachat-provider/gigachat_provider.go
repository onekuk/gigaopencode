package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gitverse.ru/kmpavloff/openai-provider-gigachat/gigachat"
	"gitverse.ru/kmpavloff/openai-provider-gigachat/openai"
)

type GigaChatProvider struct {
	baseURL      string
	httpClient   *http.Client
	tokenManager *TokenManager
	logger       *Logger
	mu           sync.Mutex
}

func NewGigaChatProvider(tokenManager *TokenManager, logger *Logger) *GigaChatProvider {
	// Create HTTP client with TLS config that skips certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &GigaChatProvider{
		baseURL:      "https://gigachat.devices.sberbank.ru/api/v1",
		tokenManager: tokenManager,
		logger:       logger,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   5 * time.Minute,
		},
	}
}

func (p *GigaChatProvider) proxyRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	requestStart := time.Now()
	fullURL := p.baseURL + endpoint

	// Получаем актуальный access token
	p.logger.Debug("Getting access token for request to %s", endpoint)
	accessToken, err := p.tokenManager.GetAccessToken()
	if err != nil {
		p.logger.Error("Failed to get access token: %v", err)
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	var reqBody io.Reader
	var jsonData []byte
	if body != nil {
		jsonData, err = json.Marshal(body)
		if err != nil {
			p.logger.Error("Failed to marshal request body: %v", err)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)

		// Логируем запрос (маскируем токен для безопасности)
		maskedToken := maskToken(accessToken)
		p.logger.Debug("GigaChat Request: %s %s, Token: %s", method, fullURL, maskedToken)
		p.logger.Debug("GigaChat Request Body: %s", string(jsonData))
	} else {
		maskedToken := maskToken(accessToken)
		p.logger.Debug("GigaChat Request: %s %s, Token: %s", method, fullURL, maskedToken)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		p.logger.Error("GigaChat Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Используем Bearer токен для авторизации
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := p.httpClient.Do(req)
	requestDuration := time.Since(requestStart)

	if err != nil {
		p.logger.Error("GigaChat Request failed: %s %s, Duration: %v, Error: %v", method, endpoint, requestDuration, err)
		return nil, err
	}

	// Логируем ответ
	p.logger.Debug("GigaChat Request completed: %s %s, Status: %d, Duration: %v", method, endpoint, resp.StatusCode, requestDuration)

	// Если уровень DEBUG, читаем и логируем тело ответа
	if p.logger.level <= LogLevelDebug && resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			// Ограничиваем размер логируемого ответа
			respBody := string(bodyBytes)
			p.logger.Debug("GigaChat Response Body: %s", respBody)

			// Восстанавливаем тело ответа
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		} else {
			p.logger.Warn("GigaChat Failed to read response body for logging: %v", err)
		}
	}

	return resp, nil
}

// Маскирует токен для безопасного логирования
func maskToken(token string) string {
	if len(token) < 10 {
		return "***"
	}
	return token[:6] + "..." + token[len(token)-4:]
}

func (p *GigaChatProvider) CreateCompletion(ctx context.Context, req openai.CompletionRequest) (*openai.CompletionResponse, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/completions", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	gigaReq, err := gigachat.ConvertOpenAIChatRequestToGigaChat(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	resp, err := p.proxyRequest(ctx, "POST", "/chat/completions", gigaReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gigaResult gigachat.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&gigaResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := gigachat.ConvertGigaChatResponseToOpenAI(&gigaResult)
	return result, nil
}

func (p *GigaChatProvider) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (<-chan *openai.ChatCompletionStreamResponse, error) {
	gigaReq, err := gigachat.ConvertOpenAIChatRequestToGigaChat(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	resp, err := p.proxyRequest(ctx, "POST", "/chat/completions", gigaReq)
	if err != nil {
		return nil, err
	}

	ch := make(chan *openai.ChatCompletionStreamResponse, 10)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Пропускаем пустые строки
			if line == "" {
				continue
			}

			// SSE формат: "data: {...}"
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// Пропускаем [DONE] маркер
				if data == "[DONE]" {
					return
				}

				var gigaData gigachat.ChatCompletionStreamDelta
				if err := json.Unmarshal([]byte(data), &gigaData); err != nil {
					p.logger.Error("Failed to decode stream data: %v, line: %s", err, data)
					continue
				}

				openaiData := gigachat.ConvertGigaChatStreamToOpenAI(&gigaData)

				select {
				case ch <- openaiData:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			p.logger.Error("Stream scanner error: %v", err)
		}
	}()

	return ch, nil
}

func (p *GigaChatProvider) CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	gigaReq := gigachat.ConvertOpenAIEmbeddingRequestToGigaChat(&req)

	resp, err := p.proxyRequest(ctx, "POST", "/embeddings", gigaReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gigaResult gigachat.EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&gigaResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := gigachat.ConvertGigaChatEmbeddingsToOpenAI(&gigaResult)
	return result, nil
}

func (p *GigaChatProvider) ListModels(ctx context.Context) (*openai.ModelsResponse, error) {
	resp, err := p.proxyRequest(ctx, "GET", "/models", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gigaResult gigachat.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&gigaResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := gigachat.ConvertGigaChatModelsToOpenAI(&gigaResult)
	return result, nil
}

func (p *GigaChatProvider) RetrieveModel(ctx context.Context, model string) (*openai.Model, error) {
	endpoint := fmt.Sprintf("/models/%s", url.PathEscape(model))
	resp, err := p.proxyRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gigaResult gigachat.Model
	if err := json.NewDecoder(resp.Body).Decode(&gigaResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := &openai.Model{
		ID:      gigaResult.ID,
		Object:  "model",
		Created: 0,
		OwnedBy: gigaResult.OwnedBy,
	}
	return result, nil
}

func (p *GigaChatProvider) DeleteModel(ctx context.Context, model string) error {
	endpoint := fmt.Sprintf("/models/%s", url.PathEscape(model))
	resp, err := p.proxyRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p *GigaChatProvider) ListFiles(ctx context.Context) (*openai.FilesResponse, error) {
	resp, err := p.proxyRequest(ctx, "GET", "/files", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.FilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) UploadFile(ctx context.Context, file io.Reader, filename, purpose string) (*openai.FileObject, error) {

	// For file upload, we'd need to handle multipart/form-data
	// This is a simplified version
	req := map[string]interface{}{
		"filename": filename,
		"purpose":  purpose,
	}

	resp, err := p.proxyRequest(ctx, "POST", "/files", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.FileObject
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) RetrieveFile(ctx context.Context, fileID string) (*openai.FileObject, error) {
	endpoint := fmt.Sprintf("/files/%s", url.PathEscape(fileID))
	resp, err := p.proxyRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.FileObject
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) DeleteFile(ctx context.Context, fileID string) error {
	endpoint := fmt.Sprintf("/files/%s", url.PathEscape(fileID))
	resp, err := p.proxyRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p *GigaChatProvider) RetrieveFileContent(ctx context.Context, fileID string) (io.ReadCloser, error) {
	endpoint := fmt.Sprintf("/files/%s/content", url.PathEscape(fileID))
	resp, err := p.proxyRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (p *GigaChatProvider) CreateFineTuningJob(ctx context.Context, req openai.FineTuningJobRequest) (*openai.FineTuningJob, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/fine_tuning/jobs", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.FineTuningJob
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) ListFineTuningJobs(ctx context.Context, after string, limit int) (*openai.FineTuningJobsResponse, error) {
	endpoint := "/fine_tuning/jobs"
	if after != "" || limit > 0 {
		params := url.Values{}
		if after != "" {
			params.Add("after", after)
		}
		if limit > 0 {
			params.Add("limit", fmt.Sprintf("%d", limit))
		}
		endpoint += "?" + params.Encode()
	}

	resp, err := p.proxyRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.FineTuningJobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) RetrieveFineTuningJob(ctx context.Context, jobID string) (*openai.FineTuningJob, error) {
	endpoint := fmt.Sprintf("/fine_tuning/jobs/%s", url.PathEscape(jobID))
	resp, err := p.proxyRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.FineTuningJob
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CancelFineTuningJob(ctx context.Context, jobID string) (*openai.FineTuningJob, error) {
	endpoint := fmt.Sprintf("/fine_tuning/jobs/%s/cancel", url.PathEscape(jobID))
	resp, err := p.proxyRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.FineTuningJob
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CreateImage(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/images/generations", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CreateImageEdit(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/images/edits", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CreateImageVariation(ctx context.Context, req openai.ImageGenerationRequest) (*openai.ImageResponse, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/images/variations", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CreateTranscription(ctx context.Context, req openai.AudioTranscriptionRequest) (*openai.AudioTranscriptionResponse, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/audio/transcriptions", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.AudioTranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CreateTranslation(ctx context.Context, req openai.AudioTranslationRequest) (*openai.AudioTranslationResponse, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/audio/translations", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.AudioTranslationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (p *GigaChatProvider) CreateSpeech(ctx context.Context, model, voice, input string) (io.ReadCloser, error) {
	req := map[string]interface{}{
		"model": model,
		"voice": voice,
		"input": input,
	}

	resp, err := p.proxyRequest(ctx, "POST", "/audio/speech", req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (p *GigaChatProvider) CreateModeration(ctx context.Context, req openai.ModerationRequest) (*openai.ModerationResponse, error) {
	resp, err := p.proxyRequest(ctx, "POST", "/moderations", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result openai.ModerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
