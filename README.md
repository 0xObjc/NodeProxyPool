# ProxyPool

本地动态代理池，基于 Gin + Mihomo。当前实现支持订阅抓取、节点池、动态端口分配、TTL 清理、健康检查，以及基础 API（`/api/getProxy` 等）。

## 快速启动

1. 安装 Go（建议 1.21+）。如果使用 Go 官方 toolchain 失败，重装/修复本机 Go 环境后再继续。
2. 修改 `configs/config.yaml`，确认订阅源 `enabled: true` 且 URL 正确。
3. 启动服务：
   ```bash
   go build ./cmd/server
   ./server -c configs/config.yaml
   ```
4. 验证：
   ```bash
   curl http://localhost:8080/health
   ./test-api.sh
   ```

## 注意事项

- Mihomo 多端口监听已接入 `WithSpecialProxy`，每个实例绑定到对应节点。暂未支持 `mixed` 协议。
- TTL 自动清理默认每 30 秒扫描一次。
- 如需更新订阅，调用 `POST /api/subscription/update`。
