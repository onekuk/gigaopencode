package openai

type Message struct {
	Role       string      `json:"role,omitempty"`
	Content    interface{} `json:"content,omitempty"`
	Name       string      `json:"name,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	Index    *int         `json:"index,omitempty"`
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []Message              `json:"messages"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	N                *int                   `json:"n,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Stop             interface{}            `json:"stop,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float64     `json:"logit_bias,omitempty"`
	User             string                 `json:"user,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	StreamOptions    *StreamOptions         `json:"stream_options,omitempty"`
	ResponseFormat   interface{}            `json:"response_format,omitempty"`
	Seed             *int                   `json:"seed,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	LogProbs     *struct {
		Content []struct {
			Token   string `json:"token"`
			LogProb float64 `json:"logprob"`
			Bytes   []int  `json:"bytes,omitempty"`
		} `json:"content"`
	} `json:"logprobs,omitempty"`
}

type ChatCompletionResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint,omitempty"`
}

type StreamChoice struct {
	Index        int     `json:"index"`
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason,omitempty"`
	LogProbs     *struct {
		Content []struct {
			Token   string `json:"token"`
			LogProb float64 `json:"logprob"`
			Bytes   []int  `json:"bytes,omitempty"`
		} `json:"content"`
	} `json:"logprobs,omitempty"`
}

type ChatCompletionStreamResponse struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	Choices           []StreamChoice `json:"choices"`
	Usage             *Usage         `json:"usage,omitempty"`
	SystemFingerprint string         `json:"system_fingerprint,omitempty"`
}

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     *int        `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

type CompletionRequest struct {
	Model            string                 `json:"model"`
	Prompt           interface{}            `json:"prompt"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	Temperature      *float64               `json:"temperature,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	N                *int                   `json:"n,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	LogProbs         *int                   `json:"logprobs,omitempty"`
	Echo             bool                   `json:"echo,omitempty"`
	Stop             interface{}            `json:"stop,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	BestOf           *int                   `json:"best_of,omitempty"`
	LogitBias        map[string]float64     `json:"logit_bias,omitempty"`
	User             string                 `json:"user,omitempty"`
	Suffix           string                 `json:"suffix,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type CompletionChoice struct {
	Text         string  `json:"text"`
	Index        int     `json:"index"`
	LogProbs     *struct {
		Tokens        []string             `json:"tokens"`
		TokenLogProbs []float64            `json:"token_logprobs"`
		TopLogProbs   []map[string]float64 `json:"top_logprobs"`
		TextOffset    []int                `json:"text_offset"`
	} `json:"logprobs,omitempty"`
	FinishReason string `json:"finish_reason"`
}

type CompletionResponse struct {
	ID                string             `json:"id"`
	Object            string             `json:"object"`
	Created           int64              `json:"created"`
	Model             string             `json:"model"`
	Choices           []CompletionChoice `json:"choices"`
	Usage             Usage              `json:"usage"`
	SystemFingerprint string             `json:"system_fingerprint,omitempty"`
}

type ImageGenerationRequest struct {
	Prompt         string  `json:"prompt"`
	Model          string  `json:"model,omitempty"`
	N              *int    `json:"n,omitempty"`
	Quality        string  `json:"quality,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Size           string  `json:"size,omitempty"`
	Style          string  `json:"style,omitempty"`
	User           string  `json:"user,omitempty"`
}

type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type ImageResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

type AudioTranscriptionRequest struct {
	File           interface{} `json:"file"`
	Model          string      `json:"model"`
	Language       string      `json:"language,omitempty"`
	Prompt         string      `json:"prompt,omitempty"`
	ResponseFormat string      `json:"response_format,omitempty"`
	Temperature    *float64    `json:"temperature,omitempty"`
	TimestampGranularities []string `json:"timestamp_granularities,omitempty"`
}

type AudioTranscriptionResponse struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Words    []struct {
		Word  string  `json:"word"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"words,omitempty"`
	Segments []struct {
		ID               int     `json:"id"`
		Seek             int     `json:"seek"`
		Start            float64 `json:"start"`
		End              float64 `json:"end"`
		Text             string  `json:"text"`
		Tokens           []int   `json:"tokens"`
		Temperature      float64 `json:"temperature"`
		AvgLogProb       float64 `json:"avg_logprob"`
		CompressionRatio float64 `json:"compression_ratio"`
		NoSpeechProb     float64 `json:"no_speech_prob"`
	} `json:"segments,omitempty"`
}

type AudioTranslationRequest struct {
	File           interface{} `json:"file"`
	Model          string      `json:"model"`
	Prompt         string      `json:"prompt,omitempty"`
	ResponseFormat string      `json:"response_format,omitempty"`
	Temperature    *float64    `json:"temperature,omitempty"`
}

type AudioTranslationResponse struct {
	Text string `json:"text"`
}

type AudioSpeechRequest struct {
	Model          string   `json:"model"`
	Input          string   `json:"input"`
	Voice          string   `json:"voice"`
	ResponseFormat string   `json:"response_format,omitempty"`
	Speed          *float64 `json:"speed,omitempty"`
}

type ModerationRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"`
}

type ModerationResult struct {
	Flagged        bool               `json:"flagged"`
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
}

type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

type FileObject struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
	Status    string `json:"status,omitempty"`
	StatusDetails string `json:"status_details,omitempty"`
}

type FilesResponse struct {
	Object string       `json:"object"`
	Data   []FileObject `json:"data"`
}

type FileDeleteResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

type FineTuningJobRequest struct {
	TrainingFile   string                 `json:"training_file"`
	ValidationFile string                 `json:"validation_file,omitempty"`
	Model          string                 `json:"model"`
	Hyperparameters map[string]interface{} `json:"hyperparameters,omitempty"`
	Suffix         string                 `json:"suffix,omitempty"`
	Integrations   []interface{}          `json:"integrations,omitempty"`
	Seed           *int                   `json:"seed,omitempty"`
}

type FineTuningJob struct {
	ID              string                 `json:"id"`
	Object          string                 `json:"object"`
	CreatedAt       int64                  `json:"created_at"`
	FinishedAt      *int64                 `json:"finished_at,omitempty"`
	Model           string                 `json:"model"`
	FineTunedModel  string                 `json:"fine_tuned_model,omitempty"`
	OrganizationID  string                 `json:"organization_id"`
	Status          string                 `json:"status"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	TrainingFile    string                 `json:"training_file"`
	ValidationFile  string                 `json:"validation_file,omitempty"`
	ResultFiles     []string               `json:"result_files"`
	TrainedTokens   *int                   `json:"trained_tokens,omitempty"`
	Error           *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Param   string `json:"param,omitempty"`
	} `json:"error,omitempty"`
	Integrations []interface{} `json:"integrations,omitempty"`
	Seed         *int          `json:"seed,omitempty"`
}

type FineTuningJobsResponse struct {
	Object  string          `json:"object"`
	Data    []FineTuningJob `json:"data"`
	HasMore bool            `json:"has_more"`
}

type BatchRequest struct {
	InputFileID      string                 `json:"input_file_id"`
	Endpoint         string                 `json:"endpoint"`
	CompletionWindow string                 `json:"completion_window"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type BatchRequestCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type Batch struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Endpoint          string                 `json:"endpoint"`
	Errors            *struct {
		Object string `json:"object"`
		Data   []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Param   string `json:"param,omitempty"`
			Line    *int   `json:"line,omitempty"`
		} `json:"data"`
	} `json:"errors,omitempty"`
	InputFileID       string                 `json:"input_file_id"`
	CompletionWindow  string                 `json:"completion_window"`
	Status            string                 `json:"status"`
	OutputFileID      string                 `json:"output_file_id,omitempty"`
	ErrorFileID       string                 `json:"error_file_id,omitempty"`
	CreatedAt         int64                  `json:"created_at"`
	InProgressAt      *int64                 `json:"in_progress_at,omitempty"`
	ExpiresAt         *int64                 `json:"expires_at,omitempty"`
	FinalizingAt      *int64                 `json:"finalizing_at,omitempty"`
	CompletedAt       *int64                 `json:"completed_at,omitempty"`
	FailedAt          *int64                 `json:"failed_at,omitempty"`
	ExpiredAt         *int64                 `json:"expired_at,omitempty"`
	CancellingAt      *int64                 `json:"cancelling_at,omitempty"`
	CancelledAt       *int64                 `json:"cancelled_at,omitempty"`
	RequestCounts     BatchRequestCounts     `json:"request_counts"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type BatchesResponse struct {
	Object  string  `json:"object"`
	Data    []Batch `json:"data"`
	FirstID string  `json:"first_id,omitempty"`
	LastID  string  `json:"last_id,omitempty"`
	HasMore bool    `json:"has_more"`
}

type ErrorResponse struct {
	Error struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param,omitempty"`
		Code    interface{} `json:"code,omitempty"`
	} `json:"error"`
}