package skills

import (
	"chat/addition/web"
	"fmt"
)

func executeWebSearch(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("missing query parameter")
	}
	result, err := web.GenerateSearchResult(query)
	if err != nil {
		return fmt.Sprintf("搜索失败: %s", err.Error()), nil
	}
	if result == "" {
		return "搜索未返回结果，请尝试其他关键词", nil
	}
	return result, nil
}
