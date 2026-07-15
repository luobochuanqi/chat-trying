package manager

import (
	"chat/adapter/volcengine"
	"chat/auth"
	"chat/utils"
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

	if user.GetDrawCount(db) <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "no draw quota remaining"})
		return
	}

	endpoint := "https://ark.cn-beijing.volces.com"
	apiKey := utils.GetStringConfs("volcengine.api_key")

	inst := volcengine.NewInstance(endpoint, apiKey)
	imageUrl, err := inst.CreateImageRequest(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	localPath := utils.StoreImage(imageUrl)
	user.DecreaseDrawCount(db, 1)

	c.JSON(http.StatusOK, gin.H{
		"url":    localPath,
		"prompt": prompt,
	})
}
