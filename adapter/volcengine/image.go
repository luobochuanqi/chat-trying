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
	truncated := prompt
	if len(truncated) > 120 {
		truncated = truncated[:120] + "..."
	}
	globals.Info(fmt.Sprintf("[seedream] request model=%s size=1K prompt=%s", SeedreamModel, truncated))

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
		globals.Warn(fmt.Sprintf("[seedream] request failed err=%v", err))
		return "", fmt.Errorf("seedream request failed: %v", err)
	}

	data := utils.MapToStruct[ImageGenerationResponse](res)
	if data == nil {
		globals.Warn(fmt.Sprintf("[seedream] parse failed"))
		return "", fmt.Errorf("seedream: cannot parse response")
	}
	if data.Error != nil && data.Error.Message != "" {
		globals.Warn(fmt.Sprintf("[seedream] api error message=%s", data.Error.Message))
		return "", fmt.Errorf("seedream error: %s", data.Error.Message)
	}
	if len(data.Data) == 0 {
		globals.Warn(fmt.Sprintf("[seedream] no image data"))
		return "", fmt.Errorf("seedream: no image data in response")
	}

	globals.Info(fmt.Sprintf("[seedream] success image_url=%s", data.Data[0].Url))
	return data.Data[0].Url, nil
}
