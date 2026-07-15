package volcengine

import (
	"fmt"
)

type ImageGenerationRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size,omitempty"`
	N              int    `json:"n,omitempty"`
	Stream         bool   `json:"stream,omitempty"`
	OutputFormat   string `json:"output_format,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Watermark      bool   `json:"watermark,omitempty"`
}

type ImageGenerationResponse struct {
	ID   string `json:"id"`
	Data []struct {
		Url           string `json:"url"`
		RevisedPrompt string `json:"revised_prompt,omitempty"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type Instance struct {
	Endpoint string
	ApiKey   string
}

func (c *Instance) GetEndpoint() string {
	return c.Endpoint
}

func (c *Instance) GetApiKey() string {
	return c.ApiKey
}

func (c *Instance) GetHeader() map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", c.GetApiKey()),
	}
}

func NewInstance(endpoint, apiKey string) *Instance {
	return &Instance{
		Endpoint: endpoint,
		ApiKey:   apiKey,
	}
}
