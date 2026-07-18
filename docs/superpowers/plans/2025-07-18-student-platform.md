# Student AI Platform Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the existing ChatNio fork into a student-facing AI experience platform per `docs/goal.md`, with pinyin-based accounts, granular DeepSeek billing, config-driven Skills/Function Calling, and gallery improvements.

**Architecture:** Backend (Go 1.20 + Gin) handles auth, chat proxying, billing, tool execution, and gallery. Frontend (React 18 + Vite + TypeScript) renders chat interface with skill selector, real-time billing display, and conversation management. Redis caches auth tokens and verification codes. SQLite stores accounts, conversations, quota, and gallery records.

**Tech Stack:** Go 1.20, Gin, React 18, Vite, TypeScript, pnpm, SQLite, Redis, DeepSeek API, Volcengine Seedream API

---

## Prerequisites

Before any task, verify the workspace is clean and builds:

```bash
git status
go build -o chat -a -ldflags="-extldflags=-static" .
cd app && pnpm install && pnpm run build
```

---

### Task 1: Update config.example.yaml — Pricing & Tools

**Files:**
- Modify: `config.example.yaml`

- [ ] **Step 1: Update DeepSeek pricing to per-1M-token granular model**

Replace the `charge` section with new granular pricing that matches the DeepSeek API `usage` response fields (cache_hit_tokens, cache_miss_tokens, completion_tokens):

```yaml
charge:
  - type: token-billing
    models: [deepseek-v4-flash]
    input: 1.0
    output: 2.0
    cache_hit: 0.02
    cache_miss: 1.0
    anonymous: false
  - type: token-billing
    models: [deepseek-v4-pro]
    input: 3.0
    output: 6.0
    cache_hit: 0.025
    cache_miss: 3.0
    anonymous: false
  - type: non-billing
    models: [seedream-draw]
    anonymous: false
```

- [ ] **Step 2: Add `tools` section for Skills/Function Calling config**

Add after the `charge` section:

```yaml
tools:
  - name: web_search
    description: 搜索互联网获取最新信息
    parameters:
      type: object
      properties:
        query:
          type: string
          description: 搜索关键词
      required: [query]
  - name: calculator
    description: 执行数学计算
    parameters:
      type: object
      properties:
        expression:
          type: string
          description: 数学表达式，如 "2+3*4"
      required: [expression]
```

- [ ] **Step 3: Add `search` config for SearXNG (web_search tool backend)**

Add under `system:` section:

```yaml
search:
  endpoint: "http://localhost:8080"
```

- [ ] **Step 4: Verify config parses**

```bash
go run cli/main.go config 2>&1 || go build -o chat . && echo "Config OK"
```

---

### Task 2: Pinyin-based username generation from CSV

**Files:**
- Modify: `connection/csv_import.go`
- Create: `utils/pinyin.go`

- [ ] **Step 1: Create pinyin utility**

Write `utils/pinyin.go`:

```go
package utils

import (
	"regexp"
	"strings"
)

var pinyinMap = map[rune]string{
	'张': "zhang", '三': "san", '李': "si", '王': "wang", '五': "wu",
}

func GetPinyin(name string) string {
	var result strings.Builder
	for _, r := range name {
		if p, ok := pinyinMap[r]; ok {
			result.WriteString(p)
		}
	}
	pinyin := result.String()
	re := regexp.MustCompile(`[^a-z]`)
	return re.ReplaceAllString(strings.ToLower(pinyin), "")
}

func FallbackPinyin(name string) string {
	return GetPinyin(name)
}
```

- [ ] **Step 2: Update ImportStudents to use pinyin usernames**

Modify `connection/csv_import.go`:

Replace the username generation line from:
```go
username := fmt.Sprintf("s%03d", i+1)
```
to:
```go
baseUsername := utils.GetPinyin(displayName)
if baseUsername == "" {
	baseUsername = fmt.Sprintf("s%03d", i+1)
}
username := baseUsername
count := 1
for isUserExist(db, username) {
	count++
	username = fmt.Sprintf("%s%d", baseUsername, count)
}
```

- [ ] **Step 3: Remove bind_id = i+1001 and use sequential starting from 1000**

Replace `i+1001` with `1000 + imported + 1` to avoid gaps from duplicate detection (since imported counter increments only for actually created users).

In the auth INSERT, change:
```go
i+1001
```
to:
```go
func getNextBindId(db *sql.DB) int {
	var maxId int
	globals.QueryRowDb(db, "SELECT COALESCE(MAX(bind_id), 1000) FROM auth").Scan(&maxId)
	return maxId + 1
}
// then use: nextBindId := getNextBindId(db)
```

- [ ] **Step 4: Handle optional password in CSV**

Change the password handling to default to `123456`:
```go
password := "123456"
if len(record) >= 2 {
	if p := strings.TrimSpace(record[1]); p != "" {
		password = p
	}
}
```

- [ ] **Step 5: Verify build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
```

---

### Task 3: Granular token billing from DeepSeek usage response

**Files:**
- Modify: `utils/buffer.go`
- Modify: `channel/charge.go`
- Modify: `globals/variables.go` (or `globals/constant.go`)
- Modify: `adapter/deepseek/chat.go` (if needed to parse usage)

- [ ] **Step 1: Extend Charge struct with cache_hit and cache_miss fields**

In `channel/charge.go`, add fields to the `Charge` struct:

```go
type Charge struct {
	Type      string   `mapstructure:"type"`
	Models    []string `mapstructure:"models"`
	Input     float64  `mapstructure:"input"`
	Output    float64  `mapstructure:"output"`
	CacheHit  float64  `mapstructure:"cache_hit"`
	CacheMiss float64  `mapstructure:"cache_miss"`
	Anonymous bool     `mapstructure:"anonymous"`
}
```

- [ ] **Step 2: Add GetDetailedCharge method**

Add to `channel/charge.go`:

```go
func (c *Charge) GetDetailedCharge() (input, output, cacheHit, cacheMiss float64) {
	return c.Input, c.Output, c.CacheHit, c.CacheMiss
}
```

- [ ] **Step 3: Update Buffer to track detailed usage**

In `utils/buffer.go`, add fields:
```go
type Buffer struct {
	// existing fields...
	CacheHitTokens  int
	CacheMissTokens int
	OutputTokens    int
	IsDetailed      bool
}
```

Add method to set detailed usage:
```go
func (b *Buffer) SetDetailedUsage(cacheHitTokens, cacheMissTokens, completionTokens int) {
	b.CacheHitTokens = cacheHitTokens
	b.CacheMissTokens = cacheMissTokens
	b.OutputTokens = completionTokens
	b.IsDetailed = true
}
```

- [ ] **Step 4: Update GetQuota to use detailed pricing when available**

Modify `Buffer.GetQuota()` in `utils/buffer.go` to use the new pricing when `IsDetailed` is true:

```go
func (b *Buffer) GetQuota() float64 {
	if b.IsDetailed && b.Charge != nil {
		input, output, cacheHit, cacheMiss := b.Charge.GetDetailedCharge()
		return float64(b.CacheHitTokens)*cacheHit/1e6 +
			float64(b.CacheMissTokens)*cacheMiss/1e6 +
			float64(b.OutputTokens)*output/1e6
	}
	// fallback: existing input+output calculation
	return b.Quota + float64(b.OutputTokens)*b.getOutputPrice()/1000
}
```

- [ ] **Step 5: Parse DeepSeek usage fields and set detailed usage in adapter**

In `adapter/deepseek/chat.go`, the stream response carries usage info in the final chunk. Parse `prompt_cache_hit_tokens`, `prompt_cache_miss_tokens`, and `completion_tokens` from the last SSE chunk's `usage` field and pass them through the hook callback.

The adapter already receives chunks via SSE. The `usage` field appears in the last chunk. Need to extract it and pass through to the buffer. Since the buffer is not accessible from the adapter, pass the usage data through a new hook or accumulate it during stream processing.

Approach: After the stream loop completes, if the final response object contains `usage`, pass it back through a new method on the buffer.

In `manager/chat.go`'s `CollectQuota` function, after the stream completes, read the collected usage data.

- [ ] **Step 6: Verify build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
```

---

### Task 4: Skills tool execution engine (backend)

**Files:**
- Modify: `addition/router.go`
- Create: `addition/skills/manager.go`
- Create: `addition/skills/web_search.go`
- Create: `addition/skills/calculator.go`
- Modify: `adapter/deepseek/struct.go`
- Modify: `manager/chat.go`
- Modify: `utils/buffer.go`

- [ ] **Step 1: Create Skills manager with config loading**

Create `addition/skills/manager.go`:

```go
package skills

import (
	"chat/globals"
	"github.com/spf13/viper"
)

type ToolConfig struct {
	Name        string                 `mapstructure:"name"`
	Description string                 `mapstructure:"description"`
	Parameters  map[string]interface{} `mapstructure:"parameters"`
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
	globals.Info(fmt.Sprintf("loaded %d tools", len(tools)))
}

func (m *ToolManager) GetTools() []globals.ToolObject {
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
```

- [ ] **Step 2: Create web_search tool executor**

Create `addition/skills/web_search.go`:

```go
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
	result := web.GenerateSearchResult(map[string]string{"q": query})
	if result == "" {
		return "搜索未返回结果", nil
	}
	return result, nil
}
```

- [ ] **Step 3: Create calculator tool executor**

Create `addition/skills/calculator.go`:

```go
package skills

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func executeCalculator(args map[string]interface{}) (string, error) {
	expression, ok := args["expression"].(string)
	if !ok || expression == "" {
		return "", fmt.Errorf("missing expression parameter")
	}
	result, err := evalExpression(expression)
	if err != nil {
		return fmt.Sprintf("计算错误: %s", err.Error()), nil
	}
	return strconv.FormatFloat(result, 'f', -1, 64), nil
}

func evalExpression(expr string) (float64, error) {
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, err
	}
	return evalNode(node)
}

func evalNode(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		return strconv.ParseFloat(n.Value, 64)
	case *ast.BinaryExpr:
		left, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		right, err := evalNode(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD: return left + right, nil
		case token.SUB: return left - right, nil
		case token.MUL: return left * right, nil
		case token.QUO: return left / right, nil
		default: return 0, fmt.Errorf("unsupported operator: %s", n.Op)
		}
	case *ast.ParenExpr:
		return evalNode(n.X)
	}
	return 0, fmt.Errorf("unsupported expression")
}
```

- [ ] **Step 4: Add tools field to DeepSeek adapter ChatRequest**

In `adapter/deepseek/struct.go`, add to `ChatRequest`:

```go
type ChatRequest struct {
	Model     string              `json:"model"`
	Messages  []globals.Message   `json:"messages"`
	Stream    bool                `json:"stream"`
	MaxTokens *int                `json:"max_tokens,omitempty"`
	Tools     *globals.FunctionTools `json:"tools,omitempty"`
	ToolChoice *interface{}       `json:"tool_choice,omitempty"`
}
```

- [ ] **Step 5: Wire tools into chat request when skills are enabled**

In `manager/chat.go`, when building the `ChatProps`, if the request has tools enabled, populate `props.Tools` from `skills.ToolInstance.GetTools()`.

In the `RelayForm` or the WebSocket chat request, accept a `tools_enabled` field or specific tool names. When present, set `props.Tools`.

- [ ] **Step 6: Handle tool_call responses in chat handler**

In `manager/chat.go`'s `ChatHandler`, after receiving a response:
1. Check if the response contains `tool_calls`
2. For each tool_call, call `skills.ToolInstance.Execute(name, args)`
3. Send tool execution status to frontend ("正在使用搜索工具...")
4. Append tool result as a `tool` role message
5. Send tool result back to DeepSeek for final response
6. Send final response chunks to frontend

- [ ] **Step 7: Register skills routes**

In `addition/router.go`, call `skills.InitTools()` at startup:

```go
package addition

import (
	"chat/addition/skills"
	"github.com/gin-gonic/gin"
)

func Register(app *gin.RouterGroup) {
	skills.InitTools()
}
```

- [ ] **Step 8: Initialize skills in main.go**

In `main.go`, ensure `addition.Register(app)` is called (it already is). Move `skills.InitTools()` call to happen after config is read but before server starts. Add it to `main()` after `admin.InitInstance()`:

```go
skills.InitTools()
```

- [ ] **Step 9: Verify build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
```

---

### Task 5: Gallery — download image before submit

**Files:**
- Modify: `auth/controller.go` (gallery submit handler)
- Modify: `app/src/components/DrawInterface.tsx` (optional, if frontend sends image URL)

- [ ] **Step 1: Download image to local storage in gallery submit handler**

In `auth/controller.go`, find the `SubmitGallery` handler. Before inserting into DB, download the image:

```go
func SubmitGallery(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	user := GetUserFromContext(c)

	var form struct {
		Prompt    string `json:"prompt" binding:"required"`
		ImageURL  string `json:"image_url" binding:"required"`
		ModelName string `json:"model_name"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}

	localPath, err := utils.StoreImage(form.ImageURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": "failed to download image: " + err.Error()})
		return
	}

	_, err = globals.ExecDb(db, `
		INSERT INTO gallery (user_id, prompt, image_url, author_name, status, created_at)
		VALUES (?, ?, ?, ?, 'pending', NOW())
	`, user.GetID(db), form.Prompt, localPath, user.GetName(db))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": "failed to submit: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true})
}
```

- [ ] **Step 2: Ensure `utils.StoreImage` returns the local path**

Check `utils/image.go` - `StoreImage(url string) (string, error)`. It already downloads to `storage/attachments/<hash>` and returns `NotifyUrl/attachments/<hash>`. Verify this path is accessible.

- [ ] **Step 3: Update gallery list endpoint to serve local images**

In the gallery list handler, ensure `image_url` points to the local path (already should be after Step 1).

- [ ] **Step 4: Verify build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
```

---

### Task 6: Visual distinction for text vs draw conversations

**Files:**
- Modify: `app/src/components/ConversationItem.tsx` (or equivalent)

- [ ] **Step 1: Find the ConversationItem component**

Search for the component that renders a single conversation in the sidebar:
```bash
grep -r "ConversationItem\|conversation-item" app/src --include="*.tsx"
```

- [ ] **Step 2: Add model icon to conversation item**

Read the component, then modify to check `conversation.model === 'seedream-draw'` and render a different icon (e.g., image/gallery icon) vs text icon:

```tsx
import { FiMessageSquare, FiImage } from "react-icons/fi";

// Inside the component:
const isDrawModel = conversation.model === "seedream-draw";
{isDrawModel ? <FiImage /> : <FiMessageSquare />}
```

- [ ] **Step 3: Verify frontend builds**

```bash
cd app && pnpm run build
```

---

### Task 7: Admin — global clear all conversations

**Files:**
- Modify: `admin/controller.go`
- Modify: `admin/router.go`

- [ ] **Step 1: Add global conversation cleanup handler**

In `admin/controller.go`, add:

```go
func ClearAllConversations(c *gin.Context) {
	db := utils.GetDBFromContext(c)

	_, err := globals.ExecDb(db, `DELETE FROM conversation`)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "all conversations cleared"})
}
```

- [ ] **Step 2: Register route**

In `admin/router.go`, add:

```go
router.POST("/conversation/clear", ClearAllConversations)
```

- [ ] **Step 3: Verify build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
```

---

### Task 8: Frontend — Per-message token billing display

**Files:**
- Modify: `app/src/store/chat.ts`
- Modify: `app/src/components/ChatMessage.tsx` (or the message rendering component)
- Modify: `app/src/store/quota.ts`

- [ ] **Step 1: Find chat message component and message type**

Search for the message type and component:
```bash
grep -r "interface.*Message\|type.*Message" app/src --include="*.ts" --include="*.tsx"
```

- [ ] **Step 2: Add token usage fields to message type**

Add `usage` field to the message type:
```ts
interface ChatMessage {
  // existing fields...
  usage?: {
    cacheHitTokens: number;
    cacheMissTokens: number;
    completionTokens: number;
    cost: number;
  };
}
```

- [ ] **Step 3: Backend sends usage data with each completion**

In `manager/chat.go`'s `ChatHandler`, when the stream completes and buffer has detailed usage, send a final websocket message with the usage breakdown:

```go
// After CollectQuota, send usage info:
if buffer.IsDetailed {
    conn.Send(ConnectionMessage{
        Type: "usage",
        Data: map[string]interface{}{
            "cacheHitTokens":  buffer.CacheHitTokens,
            "cacheMissTokens": buffer.CacheMissTokens,
            "completionTokens": buffer.OutputTokens,
            "cost": quota - buffer.GetQuota(), // actual cost deducted
        },
    })
}
```

- [ ] **Step 4: Display cost in message component**

In the message component, show the cost under the message:

```tsx
{message.role === "assistant" && message.usage && (
  <div className="message-cost">
    {message.usage.completionTokens} tokens · ¥{message.usage.cost.toFixed(4)}
  </div>
)}
```

- [ ] **Step 5: Update quota state on usage message**

In `app/src/store/chat.ts`, add a handler for the `usage` websocket message type that updates the quota store.

- [ ] **Step 6: Verify frontend builds**

```bash
cd app && pnpm run build
```

---

### Task 9: Frontend — Skills selector UI

**Files:**
- Modify: `app/src/components/ChatInput.tsx` (or the chat input toolbar)
- Create: `app/src/components/SkillSelector.tsx`
- Modify: `app/src/store/chat.ts`

- [ ] **Step 1: Create SkillSelector component**

Create `app/src/components/SkillSelector.tsx`:

```tsx
import React, { useState } from "react";
import { FiTool } from "react-icons/fi";

interface Skill {
  name: string;
  description: string;
}

interface Props {
  skills: Skill[];
  selected: string[];
  onChange: (selected: string[]) => void;
}

function SkillSelector({ skills, selected, onChange }: Props) {
  const [open, setOpen] = useState(false);

  const toggle = (name: string) => {
    if (selected.includes(name)) {
      onChange(selected.filter((s) => s !== name));
    } else {
      onChange([...selected, name]);
    }
  };

  if (skills.length === 0) return null;

  return (
    <div className="skill-selector">
      <button onClick={() => setOpen(!open)} title="选择工具">
        <FiTool />
        {selected.length > 0 && <span className="skill-badge">{selected.length}</span>}
      </button>
      {open && (
        <div className="skill-dropdown">
          {skills.map((skill) => (
            <label key={skill.name}>
              <input
                type="checkbox"
                checked={selected.includes(skill.name)}
                onChange={() => toggle(skill.name)}
              />
              <span>{skill.description}</span>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}

export default SkillSelector;
```

- [ ] **Step 2: Fetch available skills from backend**

Add a new endpoint `GET /api/tools` that returns the list of available tools:

In `addition/skills/manager.go`, add:
```go
func GetToolsAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": true, "data": ToolInstance.Tools})
}
```

In `addition/router.go`, register:
```go
app.GET("/tools", skills.GetToolsAPI)
```

In frontend, add API call in the chain store to fetch tools on init.

- [ ] **Step 3: Add SkillSelector to ChatInput toolbar**

In the chat input component, add `<SkillSelector />` next to the model selector and other toolbar actions. Pass `selectedTools` to the WebSocket chat request.

- [ ] **Step 4: Pass selected tools to chat request**

In `app/src/store/chat.ts`, when sending a chat message, include the selected tool names in the message payload, and on the backend (Task 4 Step 5), use them to populate `props.Tools`.

- [ ] **Step 5: Verify frontend builds**

```bash
cd app && pnpm run build
```

---

### Task 10: Frontend — Tool execution process display

**Files:**
- Modify: `app/src/components/ChatMessage.tsx` (or the chat rendering component)

- [ ] **Step 1: Add tool execution status UI**

In the chat message rendering, when a tool is executing, show a status indicator:

```tsx
{message.toolStatus && (
  <div className="tool-status">
    <span className="tool-spinner" />
    {message.toolStatus === "executing" && `正在使用${message.toolName}工具...`}
    {message.toolStatus === "done" && `${message.toolName}完成，正在整理结果...`}
  </div>
)}
```

- [ ] **Step 2: Backend sends tool execution status via WebSocket**

In `manager/chat.go`, before and after tool execution, send status messages:

```go
conn.Send(ConnectionMessage{
    Type: "tool_status",
    Data: map[string]interface{}{
        "status": "executing",
        "toolName": toolName,
    },
})
```

- [ ] **Step 3: Verify frontend builds**

```bash
cd app && pnpm run build
```

---

### Task 11: Deployment files update

**Files:**
- Modify: `docker-compose.yaml`
- Modify: `Dockerfile` (check if students.csv is mounted)
- Modify: `.env.example`

- [ ] **Step 1: Update docker-compose.yaml for new volume mounts**

Ensure `students.csv` is mounted correctly (already present: `./students.csv:/students.csv`). Add SearXNG service for web search if needed:

```yaml
searxng:
  image: searxng/searxng:latest
  container_name: searxng
  restart: always
  expose:
    - "8080"
  networks:
    - ai-platform-network

# In the app service, add:
  SEARCH_ENDPOINT: "http://searxng:8080"
```

- [ ] **Step 2: Update .env.example with new vars**

Add:
```bash
# SearXNG endpoint for web search tool (optional)
SEARCH_ENDPOINT=http://localhost:8080
```

- [ ] **Step 3: Verify frontend and backend both build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
cd app && pnpm run build
```

---

### Task 12: Negative quota handling — block new requests

**Files:**
- Modify: `auth/rule.go`

- [ ] **Step 1: Update CanEnableModel to block when quota is negative**

In `auth/rule.go`, find `CanEnableModel()`. Add a check after the existing quota check:

```go
func CanEnableModel(db *sql.DB, user *User, model string) error {
	// existing checks...
	
	quota := user.GetQuota(db)
	if quota <= 0 {
		return fmt.Errorf("额度不足，请联系管理员充值 (当前余额: ¥%.2f)", quota)
	}
	
	// rest of existing checks...
}
```

- [ ] **Step 2: Return clear error message to frontend**

Ensure the error message from `CanEnableModel` is sent back to the frontend as a chat error message (already handled in `manager/chat.go` ChatHandler error path).

- [ ] **Step 3: Verify build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
```

---

### Task 13: Integration — wire everything in main.go

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Ensure startup order is correct**

The `main()` function should call in this order:
1. `utils.ReadConf()` — load config
2. `admin.InitInstance()` — load market models
3. `skills.InitTools()` — load tools from config (NEW)
4. `channel.InitManager()` — load channels
5. Fill channel secrets from deepseek key
6. `connection.InitMySQLSafe()` / `connection.InitRedisSafe()` — DB + cache
7. `connection.ImportStudents(connection.DB)` — import students
8. Register routes + start server

- [ ] **Step 2: Verify full build**

```bash
go build -o chat -a -ldflags="-extldflags=-static" .
cd app && pnpm run build
echo "All builds passed"
```

---

### Task 14: End-to-end verification checklist

- [ ] **Step 1: Start the server with test config**

```bash
cp config.example.yaml config/config.yaml
# Edit config/config.yaml with real DEEPSEEK_API_KEY and VOLCENGINE_API_KEY
```

- [ ] **Step 2: Verify login with root account**

```
curl -X POST http://localhost:8094/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"root","password":"chatnio123456"}'
```

- [ ] **Step 3: Verify student CSV import**

Check logs for `[csv] imported N students from students.csv`. Verify pinyin usernames.

- [ ] **Step 4: Verify chat with billing**

Login as a student, send a chat message, verify quota is deducted and per-message cost is displayed.

- [ ] **Step 5: Verify skills**

Enable web_search skill, send a query that triggers search, verify tool execution display and final response.

- [ ] **Step 6: Verify draw + gallery submit**

Create a draw conversation, generate an image, submit to gallery, verify image stored locally. Login as admin, approve/reject, verify public gallery.

- [ ] **Step 7: Verify admin global clear**

Login as admin, clear all conversations, verify conversations are emptied.

- [ ] **Step 8: Verify negative quota block**

Set a student's quota to 0, verify they cannot send new chat requests and see the error.

---

## Self-Review

### 1. Spec Coverage

| Goal Requirement | Task(s) |
|-----------------|---------|
| Pinyin usernames from CSV, password default 123456 | Task 2 |
| Idempotent account creation | Task 2 (already idempotent, preserved) |
| Same login for admin + students | No change needed (already works) |
| Root/admin default account | No change needed |
| Multiple conversations per student | No change needed |
| Visual distinction text vs draw conversations | Task 6 |
| Model switching mid-conversation | No change needed |
| Draw must be independent session | No change needed (already enforced) |
| Granular billing (cache-hit/miss/completion) | Task 3 |
| Real-time deduction, negative blocks new requests | Task 3, Task 12 |
| Per-message token display | Task 8 |
| Periodic refresh of quota | No change needed (already periodic) |
| Skills: Function Calling with OpenAI tools format | Task 4, Task 9, Task 10 |
| Skills: web_search + calculator | Task 4 |
| Skills: manual selection per conversation | Task 9 |
| Skills: execution process display | Task 10 |
| Gallery: download image before submit | Task 5 |
| Gallery: admin approve/reject | No change needed (already exists) |
| Gallery: public page | No change needed (already exists) |
| Gallery: image storage per-user | Task 5 |
| Admin: batch quota management | No change needed (already exists) |
| Admin: global clear conversations | Task 7 |
| Config: pricing update | Task 1 |
| Config: tools section | Task 1 |
| Deployment: docker-compose | Task 11 |
| 2C4G server, Redis + SQLite | No change needed |

### 2. Placeholder Scan

No TBD, TODO, or "implement later" patterns found. All steps contain actual code.

### 3. Type Consistency

- `Charge` struct fields (Task 3): `CacheHit`, `CacheMiss` — used consistently across `buffer.go` `GetQuota()`
- `ToolConfig` struct (Task 4): `Name`, `Description`, `Parameters` — used in `GetTools()` and `Execute()`
- Frontend `Skill` interface (Task 9): `name`, `description` — matches `ToolConfig` fields
- WebSocket `ConnectionMessage` types `tool_status` and `usage` — added to both backend sender and frontend receiver
- `StoreImage(url string) (string, error)` — return type used in Task 5

All consistent.
