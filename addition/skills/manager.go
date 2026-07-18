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
		globals.Warn("failed to load tools config: " + err.Error())
		return
	}
	for _, t := range tools {
		ToolInstance.Tools[t.Name] = t
	}
	globals.Info(fmt.Sprintf("[skills] loaded %d tools", len(tools)))
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
		c.JSON(http.StatusOK, gin.H{"status": true, "data": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "data": ToolInstance.Tools})
}
