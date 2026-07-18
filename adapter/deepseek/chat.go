package deepseek

import (
	adaptercommon "chat/adapter/common"
	"chat/globals"
	"chat/utils"
	"errors"
	"fmt"
)

type ChatInstance struct {
	Endpoint         string
	ApiKey           string
	isFirstReasoning bool
	isReasonOver     bool
}

func (c *ChatInstance) GetEndpoint() string {
	return c.Endpoint
}

func (c *ChatInstance) GetApiKey() string {
	return c.ApiKey
}

func (c *ChatInstance) GetHeader() map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", c.GetApiKey()),
	}
}

func NewChatInstance(endpoint, apiKey string) *ChatInstance {
	return &ChatInstance{
		Endpoint:         endpoint,
		ApiKey:           apiKey,
		isFirstReasoning: true,
	}
}

func NewChatInstanceFromConfig(conf globals.ChannelConfig) adaptercommon.Factory {
	return NewChatInstance(
		conf.GetEndpoint(),
		conf.GetRandomSecret(),
	)
}

func (c *ChatInstance) GetChatEndpoint() string {
	return fmt.Sprintf("%s/chat/completions", c.GetEndpoint())
}

func (c *ChatInstance) GetChatBody(props *adaptercommon.ChatProps, stream bool) interface{} {
	messages := props.Message
	if len(messages) > 0 && messages[0].Role == globals.Assistant {
		messages = make([]globals.Message, len(props.Message))
		copy(messages, props.Message)
		messages[0].Role = globals.User
	}

	toolsCount := 0
	if props.Tools != nil {
		toolsCount = len(*props.Tools)
	}
	globals.Info(fmt.Sprintf("[deepseek] request model=%s messages=%d stream=%v tools=%d max_tokens=%v",
		props.Model, len(messages), stream, toolsCount,
		func() string { if props.MaxTokens != nil { return fmt.Sprintf("%d", *props.MaxTokens) } else { return "nil" } }()))

	body := ChatRequest{
		Model:            props.Model,
		Messages:         messages,
		MaxTokens:        props.MaxTokens,
		Stream:           stream,
		Temperature:      props.Temperature,
		TopP:             props.TopP,
		PresencePenalty:  props.PresencePenalty,
		FrequencyPenalty: props.FrequencyPenalty,
		Tools:            props.Tools,
		ToolChoice:       props.ToolChoice,
	}

	if stream {
		body.StreamOptions = &StreamOptions{IncludeUsage: true}
	}

	return body
}

func processChatResponse(data string) *ChatResponse {
	if form := utils.UnmarshalForm[ChatResponse](data); form != nil {
		return form
	}
	return nil
}

func processChatStreamResponse(data string) *ChatStreamResponse {
	if form := utils.UnmarshalForm[ChatStreamResponse](data); form != nil {
		return form
	}
	return nil
}

func processChatErrorResponse(data string) *ChatStreamErrorResponse {
	if form := utils.UnmarshalForm[ChatStreamErrorResponse](data); form != nil {
		return form
	}
	return nil
}

func (c *ChatInstance) ProcessLine(data string) (string, error) {
	if form := processChatStreamResponse(data); form != nil {
		if len(form.Choices) == 0 {
			return "", nil
		}

		delta := form.Choices[0].Delta

		if c.isFirstReasoning == false && !c.isReasonOver && delta.ReasoningContent == nil {
			c.isReasonOver = true
			if delta.Content != "" {
				return fmt.Sprintf("\n</think>\n\n%s", delta.Content), nil
			}
			return "\n</think>\n\n", nil
		}

		if delta.ReasoningContent != nil {
			content := *delta.ReasoningContent
			if c.isFirstReasoning {
				c.isFirstReasoning = false
				return fmt.Sprintf("<think>\n%s", content), nil
			}
			return content, nil
		}

		return delta.Content, nil
	}

	if form := processChatErrorResponse(data); form != nil {
		if form.Error.Message != "" {
			return "", errors.New(fmt.Sprintf("deepseek error: %s", form.Error.Message))
		}
	}

	return "", nil
}

func (c *ChatInstance) CreateChatRequest(props *adaptercommon.ChatProps) (string, error) {
	globals.Info(fmt.Sprintf("[deepseek] non-stream request model=%s messages=%d", props.Model, len(props.Message)))

	res, err := utils.Post(
		c.GetChatEndpoint(),
		c.GetHeader(),
		c.GetChatBody(props, false),
		props.Proxy,
	)

	if err != nil || res == nil {
		globals.Warn(fmt.Sprintf("[deepseek] non-stream request failed model=%s err=%v", props.Model, err))
		return "", fmt.Errorf("deepseek error: %s", err.Error())
	}

	data := utils.MapToStruct[ChatResponse](res)
	if data == nil {
		globals.Warn(fmt.Sprintf("[deepseek] non-stream parse failed model=%s", props.Model))
		return "", fmt.Errorf("deepseek error: cannot parse response")
	}

	if data.Usage.PromptTokens > 0 {
		globals.Info(fmt.Sprintf(
			"[deepseek] non-stream complete model=%s prompt=%d completion=%d total=%d",
			props.Model,
			data.Usage.PromptTokens,
			data.Usage.CompletionTokens,
			data.Usage.TotalTokens,
		))
	}

	if len(data.Choices) == 0 {
		return "", fmt.Errorf("deepseek error: no choices")
	}

	message := data.Choices[0].Message
	content := message.Content
	if message.ReasoningContent != nil {
		content = fmt.Sprintf("<think>\n%s\n</think>\n\n%s", *message.ReasoningContent, content)
	}

	return content, nil
}

func (c *ChatInstance) CreateStreamChatRequest(props *adaptercommon.ChatProps, callback globals.Hook) error {
	c.isFirstReasoning = true
	c.isReasonOver = false
	var lastUsage *ChatStreamUsage
	var accumulatedToolCalls globals.ToolCalls
	err := utils.EventScanner(&utils.EventScannerProps{
		Method:  "POST",
		Uri:     c.GetChatEndpoint(),
		Headers: c.GetHeader(),
		Body:    c.GetChatBody(props, true),
		Callback: func(data string) error {
			form := processChatStreamResponse(data)

			if form != nil {
				if form.Usage != nil {
					lastUsage = form.Usage
				}

				var finishReason string
				if len(form.Choices) > 0 {
					delta := form.Choices[0].Delta
					if delta.ToolCalls != nil {
						for _, chunk := range *delta.ToolCalls {
							found := false
							if chunk.Id != "" {
								for j := range accumulatedToolCalls {
									if accumulatedToolCalls[j].Id == chunk.Id {
										accumulatedToolCalls[j].Function.Arguments += chunk.Function.Arguments
										found = true
										break
									}
								}
							}
							if !found && chunk.Index != nil {
								idx := *chunk.Index
								for idx >= len(accumulatedToolCalls) {
									accumulatedToolCalls = append(accumulatedToolCalls, globals.ToolCall{})
								}
								if accumulatedToolCalls[idx].Id == "" {
									accumulatedToolCalls[idx] = chunk
								} else {
									accumulatedToolCalls[idx].Function.Arguments += chunk.Function.Arguments
								}
								found = true
							}
							if !found {
								if len(accumulatedToolCalls) > 0 {
									accumulatedToolCalls[len(accumulatedToolCalls)-1].Function.Arguments += chunk.Function.Arguments
								} else {
									accumulatedToolCalls = append(accumulatedToolCalls, chunk)
								}
							}
						}
					}
					finishReason = form.Choices[0].FinishReason
				}

				partial, err := c.ProcessLine(data)
				if err != nil {
					return err
				}

				chunk := &globals.Chunk{Content: partial}

				if finishReason == "tool_calls" {
					tools := make(globals.ToolCalls, len(accumulatedToolCalls))
					copy(tools, accumulatedToolCalls)
					chunk.ToolCall = &tools
				}

				return callback(chunk)
			}

			partial, err := c.ProcessLine(data)
			if err != nil {
				return err
			}
			return callback(&globals.Chunk{Content: partial})
		},
	}, props.Proxy)

	if err != nil {
		if form := processChatErrorResponse(err.Body); form != nil {
			if form.Error.Type == "" && form.Error.Message == "" {
				globals.Warn(fmt.Sprintf("[deepseek] stream error model=%s raw=%s", props.Model, err.Body))
				return errors.New(utils.ToMarkdownCode("json", err.Body))
			}
			globals.Warn(fmt.Sprintf("[deepseek] stream error model=%s type=%s message=%s", props.Model, form.Error.Type, form.Error.Message))
			return errors.New(fmt.Sprintf("deepseek error: %s (type: %s)", form.Error.Message, form.Error.Type))
		}
		globals.Warn(fmt.Sprintf("[deepseek] stream error model=%s err=%s", props.Model, err.Error.Error()))
		return err.Error
	}

	if lastUsage != nil && props.UsageCallback != nil {
		props.UsageCallback(lastUsage.PromptCacheHitTokens, lastUsage.PromptCacheMissTokens, lastUsage.CompletionTokens)
	}

	if lastUsage != nil {
		globals.Info(fmt.Sprintf(
			"[deepseek] stream complete model=%s prompt_tokens=%d cache_hit=%d cache_miss=%d completion=%d",
			props.Model,
			lastUsage.PromptTokens,
			lastUsage.PromptCacheHitTokens,
			lastUsage.PromptCacheMissTokens,
			lastUsage.CompletionTokens,
		))
	} else {
		globals.Info(fmt.Sprintf("[deepseek] stream complete model=%s (no usage data)", props.Model))
	}

	return nil
}
