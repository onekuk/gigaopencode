package gigachat

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
	Type    string `json:"type"`
}

type ModelsResponse struct {
	Data   []Model `json:"data"`
	Object string  `json:"object"`
}

type Message struct {
	Role             string        `json:"role"`
	Content          string        `json:"content"`
	FunctionsStateID string        `json:"functions_state_id,omitempty"`
	Attachments      []string      `json:"attachments,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
}

type MessageResponse struct {
	Role             string        `json:"role"`
	Content          string        `json:"content"`
	Created          int64         `json:"created,omitempty"`
	Name             string        `json:"name,omitempty"`
	FunctionsStateID string        `json:"functions_state_id,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
}

type FunctionCall struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type FunctionCallRequest struct {
	Name string `json:"name,omitempty"`
}

type CustomFunction struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Parameters       map[string]interface{} `json:"parameters"`
	FewShotExamples  []FewShotExample       `json:"few_shot_examples,omitempty"`
	ReturnParameters map[string]interface{} `json:"return_parameters,omitempty"`
}

type FewShotExample struct {
	Request string                 `json:"request"`
	Params  map[string]interface{} `json:"params"`
}

type ChatCompletionRequest struct {
	Model             string           `json:"model"`
	Messages          []Message        `json:"messages"`
	FunctionCall      interface{}      `json:"function_call,omitempty"`
	Functions         []CustomFunction `json:"functions,omitempty"`
	Temperature       *float64         `json:"temperature,omitempty"`
	TopP              *float64         `json:"top_p,omitempty"`
	Stream            bool             `json:"stream,omitempty"`
	MaxTokens         *int             `json:"max_tokens,omitempty"`
	RepetitionPenalty *float64         `json:"repetition_penalty,omitempty"`
	UpdateInterval    float64          `json:"update_interval,omitempty"`
}

type Usage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	PrecachedPromptTokens int `json:"precached_prompt_tokens,omitempty"`
	TotalTokens           int `json:"total_tokens"`
}

type Choice struct {
	Message      MessageResponse `json:"message"`
	Index        int             `json:"index"`
	FinishReason string          `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Object  string   `json:"object"`
}

type ChatCompletionStreamDelta struct {
	Choices []StreamChoice `json:"choices"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Object  string         `json:"object"`
	Usage   *Usage         `json:"usage,omitempty"`
}

type StreamChoice struct {
	Delta        MessageResponse `json:"delta"`
	Index        int             `json:"index"`
	FinishReason string          `json:"finish_reason,omitempty"`
}

type File struct {
	Bytes        int    `json:"bytes"`
	CreatedAt    int64  `json:"created_at"`
	Filename     string `json:"filename"`
	ID           string `json:"id"`
	Object       string `json:"object"`
	Purpose      string `json:"purpose"`
	AccessPolicy string `json:"access_policy,omitempty"`
}

type FilesResponse struct {
	Data []File `json:"data"`
}

type FileDeleted struct {
	ID           string `json:"id"`
	Deleted      bool   `json:"deleted"`
	AccessPolicy string `json:"access_policy,omitempty"`
}

type EmbeddingRequest struct {
	Model string      `json:"model"`
	Input interface{} `json:"input"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
	Usage     *Usage    `json:"usage,omitempty"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
}

type TokensCountRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type TokenCount struct {
	Object     string `json:"object"`
	Tokens     int    `json:"tokens"`
	Characters int    `json:"characters"`
}

type TokensCountResponse []TokenCount

type BalanceItem struct {
	Usage string `json:"usage"`
	Value int    `json:"value"`
}

type BalanceResponse struct {
	Balance []BalanceItem `json:"balance"`
}

type AICheckRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type AICheckResponse struct {
	Category    string  `json:"category"`
	Characters  int     `json:"characters"`
	Tokens      int     `json:"tokens"`
	AIIntervals [][]int `json:"ai_intervals,omitempty"`
}

type FunctionValidationResult struct {
	Status             int               `json:"status"`
	Message            string            `json:"message"`
	JSONAIRulesVersion string            `json:"json_ai_rules_version,omitempty"`
	Errors             []ValidationIssue `json:"errors,omitempty"`
	Warnings           []ValidationIssue `json:"warnings,omitempty"`
}

type ValidationIssue struct {
	Description    string `json:"description"`
	SchemaLocation string `json:"schema_location"`
}

type BatchRequest struct {
	Method string `json:"method,omitempty"`
}

type BatchRequestCounts struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

type BatchTask struct {
	ID            string             `json:"id"`
	Method        string             `json:"method"`
	RequestCounts BatchRequestCounts `json:"request_counts"`
	Status        string             `json:"status"`
	OutputFileID  string             `json:"output_file_id,omitempty"`
	CreatedAt     int64              `json:"created_at"`
	UpdatedAt     int64              `json:"updated_at"`
}

type BatchesListResponse struct {
	Batches []BatchTask `json:"batches"`
}

type BatchResponse struct {
	ID            string             `json:"id"`
	Method        string             `json:"method"`
	RequestCounts BatchRequestCounts `json:"request_counts"`
	Status        string             `json:"status"`
	CreatedAt     int64              `json:"created_at"`
	UpdatedAt     int64              `json:"updated_at"`
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
