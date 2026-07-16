# AGENTS.md — AI 创作平台 (ChatNio Fork)

## Build & verify

```bash
# Backend (needs CGO for sqlite3)
go build -o chat -a -ldflags="-extldflags=-static" .

# Frontend
cd app && pnpm install && pnpm run build
```

Backend is Go 1.20 with Gin. Frontend is React 18 + Vite + TypeScript + pnpm. Zero tests.

## Config bootstrap flow

On first startup, the app copies `config.example.yaml` → `config/config.yaml` via `utils/config.go:16-21`. **This only happens once** — if `config/config.yaml` already exists (e.g., from a previous deployment on a mounted volume), it is never overwritten. Environment variables (`viper.AutomaticEnv()`) override config values.

Key env vars: `DEEPSEEK_API_KEY`, `VOLCENGINE_API_KEY`, `SERVE_STATIC`.

## SQLite compatibility traps

When `mysql.host` is empty, the app falls back to SQLite (`connection/database.go:27`). All SQL goes through `PreflightSql()` in `globals/sql.go`, which converts MySQL syntax (ON DUPLICATE KEY, AUTO_INCREMENT, DECIMAL, VARCHAR, etc.) to SQLite equivalents.

**Every new `ON DUPLICATE KEY UPDATE` query must be added to the `batch` array in `globals/sql.go:34-76`** with its `ON CONFLICT` counterpart. If you skip this, SQLite mode will fail silently or panic at runtime.

## Import cycle constraint

`adapter/midjourney/storage.go` imports `connection`. Therefore the `connection` package **cannot import** `auth`, `channel`, `manager`, or `adapter`. All DB queries in `connection/` must use raw SQL with `globals.ExecDb/QueryDb` directly — never call user/auth methods.

## Channel system

- Channels load from `channel` key in config.yaml at startup
- **`state: true` is mandatory** — defaults to `false`, making the channel inactive
- The `secret` field is the API key. At startup, `main.go` iterates channels and fills empty secrets from `deepseek.api_key` config
- `seedream-draw` is NOT a channel — it routes through `POST /api/draw` directly

## Billing data model

The billing system checks the **`quota` column** (not `credit_money`). When creating users or setting credit via admin, both `quota` and `credit_money` must be set to the same value. See `admin/user.go:293` — `INSERT INTO quota ... quota = ?, credit_money = ?`.

The `draw_count` column is standalone (deducted by draw API, set by admin).

## DB schema gotchas

- `auth.bind_id` has a **UNIQUE constraint** — batch inserts must use distinct values, never `0` for all rows
- All tables are auto-created via `connection/database.go:ConnectDatabase()` — no manual migration needed

## Model market

Models shown in the frontend come from the `market` section in config.yaml, NOT from channels. The frontend fetches `GET /api/v1/market` which returns `admin.MarketInstance.GetModels()`.

## seedream-draw model

This is a special draw-only model. The backend `CanEnableModel` checks `draw_count` for it. The frontend `ChatWrapper.tsx` switches to `DrawInterface` instead of the chat UI when this model is selected. It sends `POST /api/draw` (REST, not WebSocket). Validation: `channel/worker.go:15` returns early for this model.

## CI

Single workflow: `.github/workflows/docker-ci.yaml`. Builds amd64 only, pushes to `ghcr.io/<owner>/chat-trying:latest`. GitHub Container Registry must be set to **public** in Package settings after first push.

## Admin account

First boot with empty DB auto-creates `root` / `chatnio123456` (`connection/database.go:104`).

## Students CSV

Format: `中文名,密码` per line. Usernames auto-generated as `s001`, `s002`, etc. `bind_id` values start at `1001`. Path in config: `student.csv`.
