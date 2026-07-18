# 目标文档 — AI 创作平台 (面向初中生)

面向约50名初中生同时体验 AI 功能的平台。通过 GHCR 部署到轻量云服务器（2C4G），使用管理员配置的 DeepSeek API Key 和 Seedream (火山引擎) API Key。每次启动时根据 `students.csv` 幂等创建账号（已存在则跳过），为每个账号配置人民币额度及生图次数限制。平台预期使用 7 天，无需复杂的长期维护策略。

---

## 1. 学生账号

### 1.1 CSV 格式
- 格式：`中文名,密码`（密码可选，不填默认 `123456`）
- 示例：
  ```
  张三,password123
  李四
  王五,abc456
  ```

### 1.2 用户名生成规则
- 中文名通过拼音库转为拼音用户名（如 `张三` → `zhangsan`）
- 拼音重名时自动追加数字后缀（`zhangwei1`, `zhangwei2`）
- 账号幂等创建：重启不丢失数据，已存在则跳过

### 1.3 登录与入口
- 学生和管理员使用同一登录页面，登录后根据角色展示不同界面
- 学生首次登录后看到**空白对话列表页**，自行新建对话
- 管理员默认账号：`root` / `chatnio123456`（首次空 DB 自动创建）

---

## 2. 对话系统

### 2.1 对话列表
- 每个学生可创建多个对话
- 基本操作：创建、删除
- 视觉区分文字对话和生图对话（图标标识）

### 2.2 文字对话
- 支持中途切换模型（flash ↔ pro）
- 计费按每次实际调用的模型分别计算
- 切换模型不影响上下文历史

### 2.3 生图对话
- 生图必须创建**独立的新对话**，选择生图模型
- 不能在文字聊天对话中切换到生图模型
- 仅使用 Seedream 单图生成模式（`n=1`）

### 2.4 对话清理
- 平台使用周期仅 7 天，无需自动清理策略
- 管理员后台保留一键清空所有对话的功能

---

## 3. 额度与计费

### 3.1 计费模型
- 基于 DeepSeek API 返回的 `usage` 字段实时扣费：
  - `prompt_cache_hit_tokens` — 缓存命中 tokens
  - `prompt_cache_miss_tokens` — 缓存未命中 tokens
  - `completion_tokens` — 输出 tokens
- 定价（每 1M tokens，单位：元）：

  | 项目 | deepseek-v4-flash | deepseek-v4-pro |
  |------|-------------------|-----------------|
  | 输入(缓存命中) | 0.02 | 0.025 |
  | 输入(缓存未命中) | 1 | 3 |
  | 输出 | 2 | 6 |

### 3.2 扣费规则
- 每次 API 调用返回后实时从余额扣除
- 允许余额变为负数（防止对话硬中断）
- 余额为负数后**禁止发起新请求**
- 学生可查看每条消息的 token 消耗明细

### 3.3 生图限额
- 按次数计算，独立于人民币额度
- 每次生图消耗 1 次，由管理员批量配置

### 3.4 余额展示
- 前端定时刷新显示当前额度和剩余生图次数

---

## 4. Skills（Function Calling 工具）

### 4.1 实现方式
- 使用 DeepSeek 原生 Function Calling（OpenAI tools 格式），**不引入 MCP 协议**
- 后端作为代理执行工具：DeepSeek 返回 tool_call → 后端执行 → 结果回传 → 模型整合回复

### 4.2 启用方式
- Students **手动选择启用**某个 Skill（不默认激活）
- UI 位置：输入框旁边，下拉菜单/开关形式选择
- 选中后本次对话的请求会携带对应 tool 定义

### 4.3 执行过程展示
- 学生可见简化版执行过程，例如 "正在使用搜索工具..." → "搜索完成，正在整理结果..."
- 目的：让初中生理解 AI 如何调用工具，而非黑盒体验

### 4.4 预置工具
在 config.yaml 中配置，初始预置两个：

| 工具名 | 功能 | 描述 |
|--------|------|------|
| `web_search` | 联网搜索 | 执行 HTTP 搜索请求获取实时信息 |
| `calculator` | 计算器 | 精确数学运算 |

### 4.5 工具定义格式
管理员在 config.yaml 的 `tools` 段定义工具，采用 OpenAI function calling 兼容格式：

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
          description: 数学表达式
      required: [expression]
```

后端启动时加载 tools 配置，请求时拼入 DeepSeek API 的 `tools` 字段。当模型返回 tool_call 时，后端按 `name` 匹配执行对应工具，结果回传模型。

---

## 5. 生图（Seedream）

### 5.1 API
- 端点：`POST https://ark.cn-beijing.volces.com/api/v3/images/generations`
- 认证：`Authorization: Bearer ${VOLCENGINE_API_KEY}`
- 同步调用模式，返回图片 URL（有效期 **24 小时**）
- 仅使用单图模式（`n=1`）

### 5.2 Gallery 提交流程
1. 学生在生图对话中，每张生成的图片旁有独立的 **"提交到 Gallery"** 按钮
2. 提交后图片进入审核队列，状态为"待审核"
3. 学生提交前需**下载图片到本地存储**（URL 24h 过期）

### 5.3 管理员审核
- 审核列表展示：缩略图、提示词、学生名、提交时间
- 操作：通过 / 驳回（无需填写理由）
- 驳回后图片**保留在磁盘**但不公开，审核列表标记"已驳回"（留档备查，不自动清理）

### 5.4 公开 Gallery
- 所有学生可访问的公开画廊页面
- 展示形式：网格布局，展示缩略图和提示词
- 只有审核通过的图片才公开显示

### 5.5 图片存储
- 路径规则：`storage/gallery/{userId}/{timestamp}_{prompt摘要}.png`
- 同时保存提示词和生成参数（JSON 配套文件）

---

## 6. 管理员功能

- 查看所有学生列表及余额 / 生图次数
- 单个或批量充值 / 扣款
- 单个或批量设置生图次数
- Gallery 审核（通过 / 驳回，无需理由）
- 一键清空所有学生对话
- 管理员与学生在同一登录入口，根据角色显示不同界面

---

## 7. 部署架构

| 组件 | 选型 |
|------|------|
| 后端 | Go 1.20 + Gin |
| 前端 | React 18 + Vite + TypeScript + pnpm |
| 数据库 | SQLite（若并发不足可切换 MySQL 镜像） |
| 缓存 | Redis（Auth token 72h TTL + 验证码缓存） |
| 部署 | Docker Compose，GHCR 拉取镜像 |

### 7.1 服务器
- 轻量云服务器 **2C4G**
- 宿主机端口 `8000` 映射容器 `8094`

### 7.2 必要环境变量
| 变量 | 说明 |
|------|------|
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥 |
| `VOLCENGINE_API_KEY` | 火山引擎 ARK API 密钥（Seedream） |
| `GHCR_USER` | GitHub 用户名（docker-compose 拉取镜像用） |
| `SERVE_STATIC` | 是否托管前端静态文件 |

### 7.3 并发评估
- DeepSeek 限流：flash 2500 RPM / pro 500 RPM — 50人场景绰绰有余
- SQLite 写入并发：若性能不足，可切换为 MySQL（代码已支持双数据库）
- 2C4G 运行 Go + Redis + SQLite 无压力

---

## 8. 未决事项

- [ ] 前端输入框旁的 Skills 选择器具体 UI 样式（实现时确定）
- [ ] Skills 工具执行后的错误处理（搜索超时、计算异常等如何向学生展示）
- [ ] 若 SQLite 并发不足，MySQL 迁移由管理员手动切换 config 即可（代码已支持双数据库，无需额外开发）
