package deepseek

import (
	"chat/globals"
)

// DeepSeek API is similar to OpenAI API with additional reasoning content

type ChatRequest struct {
	Model            string                 `json:"model"`
	Messages         []globals.Message      `json:"messages"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	Stream           bool                   `json:"stream"`
	StreamOptions    *StreamOptions         `json:"stream_options,omitempty"`
	Temperature      *float32               `json:"temperature,omitempty"`
	TopP             *float32               `json:"top_p,omitempty"`
	PresencePenalty  *float32               `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32               `json:"frequency_penalty,omitempty"`
	Tools            *globals.FunctionTools `json:"tools,omitempty"`
	ToolChoice       *interface{}           `json:"tool_choice,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// ChatResponse is the native http request body for deepseek
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int             `json:"index"`
		Message      globals.Message `json:"message"`
		FinishReason string          `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatStreamResponse is the stream response body for deepseek
type ChatStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta        globals.Message `json:"delta"`
		Index        int             `json:"index"`
		FinishReason string          `json:"finish_reason"`
	} `json:"choices"`
	Usage *ChatStreamUsage `json:"usage,omitempty"`
}

type ChatStreamUsage struct {
	PromptTokens          int `json:"prompt_tokens"`
	CompletionTokens      int `json:"completion_tokens"`
	TotalTokens           int `json:"total_tokens"`
	PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
}

type ChatStreamErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
