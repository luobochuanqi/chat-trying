package skills

import (
	"chat/globals"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ToolConfig struct {
	Name        string                `mapstructure:"name"`
	Description string                `mapstructure:"description"`
	Parameters  globals.ToolParameters `mapstructure:"parameters"`
}

var ToolInstance *ToolManager

type ToolManager struct {
	Tools map[string]ToolConfig
}

func InitTools() {
	ToolInstance = &ToolManager{Tools: make(map[string]ToolConfig)}
	var tools []ToolConfig
	if err := viper.UnmarshalKey("tools", &tools); err != nil {
		globals.Warn("[skills] failed to load tools config: " + err.Error())
		globals.Warn("[skills] please add a `tools` section to config/config.yaml (see config.example.yaml for reference)")
		return
	}
	if len(tools) == 0 {
		globals.Warn("[skills] no tools configured — please add a `tools` section to config/config.yaml")
		return
	}
	for _, t := range tools {
		ToolInstance.Tools[t.Name] = t
	}
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		names = append(names, t.Name)
	}
	globals.Info(fmt.Sprintf("[skills] loaded %d tools: %v", len(tools), names))
}

func (m *ToolManager) GetToolsForRequest(toolNames []string) []globals.ToolObject {
	if m == nil {
		return nil
	}
	if len(toolNames) == 0 {
		return m.GetAllTools()
	}
	var result []globals.ToolObject
	for _, name := range toolNames {
		if t, ok := m.Tools[name]; ok {
			result = append(result, globals.ToolObject{
				Type: "function",
				Function: globals.ToolFunction{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				},
			})
		}
	}
	return result
}

func (m *ToolManager) GetAllTools() []globals.ToolObject {
	var result []globals.ToolObject
	for _, t := range m.Tools {
		result = append(result, globals.ToolObject{
			Type: "function",
			Function: globals.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}
	return result
}

func (m *ToolManager) Execute(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "web_search":
		return executeWebSearch(args)
	case "calculator":
		return executeCalculator(args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func GetToolsAPI(c *gin.Context) {
	if ToolInstance == nil {
		globals.Info("[skills] GET /api/tools — ToolInstance is nil (InitTools not called?)")
		c.JSON(http.StatusOK, gin.H{"status": true, "data": []interface{}{}})
		return
	}
	count := len(ToolInstance.Tools)
	globals.Info(fmt.Sprintf("[skills] GET /api/tools — returning %d tools", count))
	c.JSON(http.StatusOK, gin.H{"status": true, "data": ToolInstance.Tools})
}
