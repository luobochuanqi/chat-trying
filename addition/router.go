package addition

import (
	"chat/addition/skills"

	"github.com/gin-gonic/gin"
)

func Register(app *gin.RouterGroup) {
	app.GET("/tools", skills.GetToolsAPI)
}
