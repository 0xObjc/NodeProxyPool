# Repository Guidelines

所有沟通、代码注释、文档全部使用**中文**，新文件使用 UTF-8（无 BOM）。
默认采取破坏性改动并拒绝向后兼容，主动清理过时代码、接口、文档；如无迁移需求需说明"**无迁移，直接替换**"。
回复格式必须：
- 在开头提供【前置说明】（简要说明：本次任务、假设、是否调用工具等）。
- 若有工具/MCP/外部调用，在结尾提供【工具调用简报】（列出用过哪些工具、用途和结论）。


## Project Structure & Modules
- `cmd/server`: Gin HTTP entrypoint; wires config, logging, subscription manager, health checker, proxy manager.
- `internal/config`: YAML config structs + validation.
- `internal/subscription`: fetcher, parser (supports Base64, YAML-only `proxies`), manager for periodic updates.
- `internal/node`: node model, pool, health checker.
- `internal/proxy`: port allocator, proxy manager (creates Mihomo listeners), TTL cleaner.
- `internal/mihomo`: embedded Mihomo adapter; creates per-instance socks/http listeners bound to selected node.
- `internal/api`: Gin handlers/middleware for `/api/*`.
- `configs/config.yaml`: runtime settings; ensure `subscription.sources[].enabled: true`.
- `test-api.sh`: simple API smoke script.

## Build, Test, Run
- Build: `GOCACHE=$(pwd)/.cache/go-build go build ./cmd/server`
- Run: `./server -c configs/config.yaml`
- Smoke test: `./test-api.sh` (expects server on `http://localhost:8080`).
- Clean cache (optional): `rm -rf .cache/go-build`

## Coding Style & Naming
- Go 1.24; use `gofmt` defaults (tabs, standard imports).
- Packages/files lowercase with hyphen-free names (`manager.go`, `pool.go`).
- Log with `pkg/logger` (zap). Keep comments brief and purposeful.
- Config keys are lower_snake_case in YAML.

## Testing Guidelines
- Place Go tests alongside code as `_test.go`; prefer table-driven tests.
- For HTTP handlers, use `httptest` + Gin test helpers; assert JSON responses.
- For proxy selection/TTL, unit test manager logic; mock Mihomo interactions if needed.
- No coverage bar set yet—note what’s covered in PRs.

## Commit & PR Guidelines
- Commits: short, imperative summaries (e.g., `Handle YAML-only subscriptions`).
- PRs should include: what/why, key commands run (build/tests), API/behavior impacts, and related issues. Screenshots optional; logs for failures helpful.

## Security & Configuration Tips
- Subscriptions may be Base64 or plain YAML; only `proxies` are required. If V2Ray parsing fails, YAML `proxies` are consumed directly.
- Health checker can mark nodes unavailable; proxy manager falls back to nodes even if checks fail, so prefer fixing upstream reachability.
- Default ports are 20000-30000; adjust in `configs/config.yaml` if conflicts occur.
- If `$GOCACHE` default is not writable, keep using project-local `.cache/go-build` for builds and tidy.
