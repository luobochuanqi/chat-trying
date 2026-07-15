# 部署指南

## 1. 首次 GitHub 仓库设置

### 1.1 开启 GHCR 权限

仓库 → **Settings** → **Actions** → **General** → **Workflow permissions** → 勾选 **Read and write permissions**

### 1.2 设置镜像为公开

Action 成功运行一次后，到仓库主页 → 右侧 **Packages** → 点击 `chat-trying` → **Package settings** → **Visibility** → 选 **Public**

否则服务器上 `docker pull` 会报 403。

### 1.3 推送 main 触发构建

```bash
git push origin main
```

Action 运行结束后，镜像会出现在 `ghcr.io/<你的用户名>/chat-trying:latest`

---

## 2. 服务端部署

### 2.1 准备

```bash
# 目录结构
mkdir -p /opt/ai-platform/{config,logs,storage}
cd /opt/ai-platform
```

### 2.2 docker-compose.yml

```yaml
version: '3'
services:
  redis:
    image: redis:alpine
    container_name: redis
    restart: always
    expose:
      - "6379"
    volumes:
      - ./redis:/data

  app:
    image: ghcr.io/你的GitHub用户名/chat-trying:latest
    container_name: ai-platform
    restart: always
    ports:
      - "8000:8094"
    depends_on:
      - redis
    environment:
      REDIS_HOST: redis
      REDIS_PORT: 6379
      REDIS_PASSWORD: ""
      REDIS_DB: 0
      SECRET: "替换为32位以上随机字符串"
      VOLCENGINE_API_KEY: "你的火山引擎ARK密钥"
      DEEPSEEK_API_KEY: "你的DeepSeek API密钥"
    volumes:
      - ./config:/config
      - ./logs:/logs
      - ./storage:/storage
      - ./students.csv:/students.csv
```

### 2.3 配置文件

创建 `config/config.yaml`：

```yaml
redis:
  host: redis
  port: 6379
  db: 0

secret: "与上面SECRET保持一致"

server:
  port: 8094

volcengine:
  api_key: "你的火山引擎ARK密钥"

system:
  general:
    backend: "http://localhost:8094"
  site:
    closeregister: true
  common:
    imagestore: true

student:
  initial_credit: 10.0     # 每名学生初始¥额度
  initial_draws: 50        # 每名学生初始生图次数
  csv: "/students.csv"
```

### 2.4 学生名单

创建 `students.csv`（格式：中文名,密码）：

```csv
张三,abc123
李四,abc123
王五,abc123
```

用户名自动生成为 `s001`、`s002` … `s050`。

### 2.5 启动

```bash
docker pull ghcr.io/你的GitHub用户名/chat-trying:latest
docker compose up -d
```

---

## 3. 访问

| 地址 | 说明 |
|------|------|
| `http://服务器IP:8000` | 学生登录 + 聊天 + 生图 |
| `http://服务器IP:8000/gallery` | 公开作品墙 |
| `http://服务器IP:8000/login` | 登录页 |

### 管理员

| 用户名 | 密码 | 说明 |
|--------|------|------|
| `root` | `chatnio123456` | 首次启动自动创建 |

登录后左上角菜单 → **Admin** 进入后台，可管理学生额度、审核作品。

---

## 4. 常用命令

```bash
docker compose logs -f          # 查看日志
docker compose restart          # 重启
docker compose pull app         # 拉取新镜像
docker compose up -d            # 应用更新
```
