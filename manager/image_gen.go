package manager

import (
	"chat/adapter/volcengine"
	"chat/auth"
	"chat/globals"
	"chat/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DrawForm struct {
	Prompt string `json:"prompt" binding:"required"`
}

func DrawAPI(c *gin.Context) {
	username := utils.GetUserFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var form DrawForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	prompt := strings.TrimSpace(form.Prompt)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	db := utils.GetDBFromContext(c)
	user := &auth.User{Username: username}

	drawCount := user.GetDrawCount(db)
	globals.Info(fmt.Sprintf("[draw] user=%s draw_count=%d prompt=%s", username, drawCount,
		func() string { if len(prompt) > 100 { return prompt[:100] + "..." } else { return prompt } }()))

	if drawCount <= 0 {
		globals.Warn(fmt.Sprintf("[draw] user=%s no quota", username))
		c.JSON(http.StatusForbidden, gin.H{"error": "no draw quota remaining"})
		return
	}

	endpoint := "https://ark.cn-beijing.volces.com"
	apiKey := utils.GetStringConfs("volcengine.api_key")

	inst := volcengine.NewInstance(endpoint, apiKey)
	imageUrl, err := inst.CreateImageRequest(prompt)
	if err != nil {
		globals.Warn(fmt.Sprintf("[draw] user=%s generation failed err=%v", username, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	globals.Info(fmt.Sprintf("[draw] user=%s storing image url=%s", username, imageUrl))
	localPath := utils.StoreImage(imageUrl)
	user.DecreaseDrawCount(db, 1)
	newCount := user.GetDrawCount(db)

	globals.Info(fmt.Sprintf("[draw] user=%s complete draw_count=%d->%d local=%s", username, drawCount, newCount, localPath))

	c.JSON(http.StatusOK, gin.H{
		"url":        localPath,
		"prompt":     prompt,
		"draw_count": newCount,
	})
}
