package gigachat

import (
	"encoding/json"
	"fmt"

	"gitverse.ru/kmpavloff/openai-provider-gigachat/openai"
)

func ConvertChatMessageToGigaChat(role, content, name string, attachments []string) Message {
	msg := Message{
		Role:    role,
		Content: content,
	}

	if len(attachments) > 0 {
		msg.Attachments = attachments
	}

	return msg
}

func ConvertOpenAIMessageToGigaChat(openaiMsg openai.Message) (Message, error) {
	role := openaiMsg.Role

	// GigaChat не поддерживает роль "tool", конвертируем в "function"
	if role == "tool" {
		role = "function"
	}

	msg := Message{
		Role: role,
	}

	// Конвертируем tool_calls из OpenAI в function_call для GigaChat
	if role == "assistant" && len(openaiMsg.ToolCalls) > 0 {
		// GigaChat поддерживает только один function_call, берем первый
		firstToolCall := openaiMsg.ToolCalls[0]

		// Парсим arguments из строки в map
		var args map[string]interface{}
		if firstToolCall.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(firstToolCall.Function.Arguments), &args); err != nil {
				// Если не удалось распарсить, используем пустой объект
				args = make(map[string]interface{})
			}
		} else {
			args = make(map[string]interface{})
		}

		msg.FunctionCall = &FunctionCall{
			Name:      firstToolCall.Function.Name,
			Arguments: args,
		}
	}

	switch v := openaiMsg.Content.(type) {
	case string:
		// Для function/tool сообщений контент должен быть валидным JSON
		if role == "function" {
			// Проверяем, является ли строка уже JSON
			var testJSON interface{}
			if err := json.Unmarshal([]byte(v), &testJSON); err != nil {
				// Если не JSON, оборачиваем в объект
				wrapped := map[string]string{"result": v}
				jsonData, err := json.Marshal(wrapped)
				if err != nil {
					return msg, fmt.Errorf("failed to marshal function result: %w", err)
				}
				msg.Content = string(jsonData)
			} else {
				msg.Content = v
			}
		} else {
			msg.Content = v
		}
	case []interface{}:
		contentJSON, err := json.Marshal(v)
		if err != nil {
			return msg, fmt.Errorf("failed to marshal content: %w", err)
		}
		msg.Content = string(contentJSON)
	default:
		if v != nil {
			contentJSON, err := json.Marshal(v)
			if err != nil {
				return msg, fmt.Errorf("failed to marshal content: %w", err)
			}
			msg.Content = string(contentJSON)
		}
	}

	return msg, nil
}

func ConvertOpenAIChatRequestToGigaChat(openaiReq *openai.ChatCompletionRequest) (*ChatCompletionRequest, error) {
	req := &ChatCompletionRequest{
		Model:          openaiReq.Model,
		Temperature:    openaiReq.Temperature,
		TopP:           openaiReq.TopP,
		Stream:         openaiReq.Stream,
		MaxTokens:      openaiReq.MaxTokens,
		UpdateInterval: 0,
	}

	// GigaChat требует единственное системное сообщение первым
	// Объединяем все системные сообщения в одно
	var systemContents []string
	var otherMessages []Message

	for _, openaiMsg := range openaiReq.Messages {
		msg, err := ConvertOpenAIMessageToGigaChat(openaiMsg)
		if err != nil {
			return nil, err
		}

		if msg.Role == "system" {
			systemContents = append(systemContents, msg.Content)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	// Если есть системные сообщения, объединяем их в одно и добавляем первым
	if len(systemContents) > 0 {
		var combinedContent string
		if len(systemContents) == 1 {
			combinedContent = systemContents[0]
		} else {
			// Объединяем несколько системных сообщений через двойной перевод строки
			combinedContent = ""
			for i, content := range systemContents {
				if i > 0 {
					combinedContent += "\n\n"
				}
				combinedContent += content
			}
		}

		systemMessage := Message{
			Role:    "system",
			Content: combinedContent,
		}
		req.Messages = append([]Message{systemMessage}, otherMessages...)
	} else {
		req.Messages = otherMessages
	}

	if len(openaiReq.Tools) > 0 {
		for _, tool := range openaiReq.Tools {
			function := CustomFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
			}

			if params, ok := tool.Function.Parameters.(map[string]interface{}); ok {
				function.Parameters = params
			}

			req.Functions = append(req.Functions, function)
		}
	}

	if openaiReq.ToolChoice != nil {
		switch v := openaiReq.ToolChoice.(type) {
		case string:
			req.FunctionCall = v
		case map[string]interface{}:
			if funcMap, ok := v["function"].(map[string]interface{}); ok {
				if name, ok := funcMap["name"].(string); ok {
					req.FunctionCall = FunctionCallRequest{Name: name}
				}
			}
		}
	}

	if openaiReq.FrequencyPenalty != nil {
		req.RepetitionPenalty = openaiReq.FrequencyPenalty
	}

	return req, nil
}

func ConvertGigaChatResponseToOpenAI(gigachatResp *ChatCompletionResponse) *openai.ChatCompletionResponse {
	choices := make([]openai.Choice, len(gigachatResp.Choices))

	for i, choice := range gigachatResp.Choices {
		message := openai.Message{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
		}

		if choice.Message.FunctionCall != nil {
			var argsJSON string
			switch v := choice.Message.FunctionCall.Arguments.(type) {
			case string:
				argsJSON = v
			case map[string]interface{}:
				argsJSON = mustMarshalJSON(v)
			default:
				argsJSON = mustMarshalJSON(v)
			}

			message.ToolCalls = []openai.ToolCall{
				{
					ID:   fmt.Sprintf("call_%d", i),
					Type: "function",
					Function: openai.FunctionCall{
						Name:      choice.Message.FunctionCall.Name,
						Arguments: argsJSON,
					},
				},
			}
		}

		choices[i] = openai.Choice{
			Index:        choice.Index,
			Message:      message,
			FinishReason: choice.FinishReason,
		}
	}

	return &openai.ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", gigachatResp.Created),
		Object:  "chat.completion",
		Created: gigachatResp.Created,
		Model:   gigachatResp.Model,
		Choices: choices,
		Usage: openai.Usage{
			PromptTokens:     gigachatResp.Usage.PromptTokens,
			CompletionTokens: gigachatResp.Usage.CompletionTokens,
			TotalTokens:      gigachatResp.Usage.TotalTokens,
		},
	}
}

func ConvertGigaChatStreamToOpenAI(gigachatStream *ChatCompletionStreamDelta) *openai.ChatCompletionStreamResponse {
	choices := make([]openai.StreamChoice, len(gigachatStream.Choices))

	for i, choice := range gigachatStream.Choices {
		delta := openai.Message{
			Content: choice.Delta.Content,
		}

		// GigaChat может возвращать пустую строку для role, OpenAI требует либо "assistant" либо отсутствие поля
		if choice.Delta.Role != "" {
			delta.Role = choice.Delta.Role
		}

		if choice.Delta.FunctionCall != nil {
			var argsJSON string
			switch v := choice.Delta.FunctionCall.Arguments.(type) {
			case string:
				argsJSON = v
			case map[string]interface{}:
				argsJSON = mustMarshalJSON(v)
			default:
				argsJSON = mustMarshalJSON(v)
			}

			toolCallIndex := 0
			delta.ToolCalls = []openai.ToolCall{
				{
					Index: &toolCallIndex,
					ID:    fmt.Sprintf("call_%d", i),
					Type:  "function",
					Function: openai.FunctionCall{
						Name:      choice.Delta.FunctionCall.Name,
						Arguments: argsJSON,
					},
				},
			}
		}

		choices[i] = openai.StreamChoice{
			Index: choice.Index,
			Delta: delta,
		}

		if choice.FinishReason != "" {
			choices[i].FinishReason = choice.FinishReason
		}
	}

	result := &openai.ChatCompletionStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", gigachatStream.Created),
		Object:  "chat.completion.chunk",
		Created: gigachatStream.Created,
		Model:   gigachatStream.Model,
		Choices: choices,
	}

	if gigachatStream.Usage != nil {
		result.Usage = &openai.Usage{
			PromptTokens:     gigachatStream.Usage.PromptTokens,
			CompletionTokens: gigachatStream.Usage.CompletionTokens,
			TotalTokens:      gigachatStream.Usage.TotalTokens,
		}
	}

	return result
}

func ConvertGigaChatModelsToOpenAI(gigachatModels *ModelsResponse) *openai.ModelsResponse {
	data := make([]openai.Model, len(gigachatModels.Data))

	for i, model := range gigachatModels.Data {
		data[i] = openai.Model{
			ID:      model.ID,
			Object:  "model",
			Created: 0,
			OwnedBy: model.OwnedBy,
		}
	}

	return &openai.ModelsResponse{
		Object: "list",
		Data:   data,
	}
}

func ConvertGigaChatEmbeddingsToOpenAI(gigachatEmb *EmbeddingResponse) *openai.EmbeddingResponse {
	data := make([]openai.Embedding, len(gigachatEmb.Data))

	for i, emb := range gigachatEmb.Data {
		data[i] = openai.Embedding{
			Object:    "embedding",
			Embedding: emb.Embedding,
			Index:     emb.Index,
		}
	}

	totalTokens := 0
	if len(gigachatEmb.Data) > 0 && gigachatEmb.Data[0].Usage != nil {
		totalTokens = gigachatEmb.Data[0].Usage.PromptTokens
	}

	return &openai.EmbeddingResponse{
		Object: "list",
		Data:   data,
		Model:  gigachatEmb.Model,
		Usage: openai.Usage{
			PromptTokens: totalTokens,
			TotalTokens:  totalTokens,
		},
	}
}

func ConvertOpenAIEmbeddingRequestToGigaChat(openaiReq *openai.EmbeddingRequest) *EmbeddingRequest {
	return &EmbeddingRequest{
		Model: openaiReq.Model,
		Input: openaiReq.Input,
	}
}

func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
