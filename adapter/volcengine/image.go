package volcengine

import (
	"chat/globals"
	"chat/utils"
	"fmt"
)

func (c *Instance) GetImageEndpoint() string {
	return fmt.Sprintf("%s/api/v3/images/generations", c.GetEndpoint())
}

func (c *Instance) CreateImageRequest(prompt string) (string, error) {
	res, err := utils.Post(
		c.GetImageEndpoint(),
		c.GetHeader(),
		ImageGenerationRequest{
			Model:          SeedreamModel,
			Prompt:         prompt,
			Size:           "1K",
			N:              1,
			Stream:         false,
			ResponseFormat: "url",
		},
		globals.ProxyConfig{},
	)
	if err != nil || res == nil {
		return "", fmt.Errorf("seedream request failed: %v", err)
	}

	data := utils.MapToStruct[ImageGenerationResponse](res)
	if data == nil {
		return "", fmt.Errorf("seedream: cannot parse response")
	}
	if data.Error != nil && data.Error.Message != "" {
		return "", fmt.Errorf("seedream error: %s", data.Error.Message)
	}
	if len(data.Data) == 0 {
		return "", fmt.Errorf("seedream: no image data in response")
	}

	return data.Data[0].Url, nil
}
