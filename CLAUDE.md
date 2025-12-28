# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ProxyPool is a local dynamic proxy pool built with Go, Gin, and Mihomo. It converts subscription-based proxy nodes (from airport services) into an API-driven, multi-port proxy pool with lifecycle management. The system provides dynamic port allocation, TTL-based cleanup, health checking, and intelligent node selection.

## Build and Run Commands

### Build and Start Server
```bash
# Build the server
go build ./cmd/server

# Run with default config
./server -c configs/config.yaml

# Run with custom config
./server -c /path/to/config.yaml
```

### Testing
```bash
# Health check
curl http://localhost:8080/health

# Run full API test suite
./test-api.sh

# Test specific endpoints
curl http://localhost:8080/api/nodePool
curl -X POST http://localhost:8080/api/getProxy -H "Content-Type: application/json" -d '{}'
```

### Development
```bash
# Install dependencies
go mod tidy

# Run in development mode (edit config.yaml: mode: debug)
./server -c configs/config.yaml
```

## Architecture Overview

### Core Components and Data Flow

The system follows a layered architecture with clear separation of concerns:

1. **Subscription Layer** (`internal/subscription/`)
   - Fetches proxy subscriptions from configured URLs
   - Parses subscription content using Mihomo's converter
   - Updates node pool periodically (default: 6h)
   - Entry point: `Manager.Start()` in cmd/server/main.go:56

2. **Node Pool** (`internal/node/`)
   - Central registry of all available proxy nodes
   - Provides filtering by availability, delay, type, region
   - Maintains type-based indexes for fast lookup
   - Thread-safe with RWMutex protection

3. **Proxy Management** (`internal/proxy/`)
   - **PortAllocator**: Manages dynamic port allocation (20000-30000 range)
   - **Manager**: Core orchestrator that coordinates node selection, port allocation, and listener creation
   - **TTLCleaner**: Background goroutine that scans every 30s to cleanup expired instances
   - **ProxyInstance**: Represents a single active proxy with TTL and metadata

4. **Mihomo Integration** (`internal/mihomo/`)
   - Wraps Mihomo core as an embedded library (not external process)
   - Creates independent SOCKS5/HTTP listeners per proxy instance
   - Uses `inbound.WithSpecialProxy()` to bind each listener to specific node
   - Maintains global proxy registry for routing

5. **API Layer** (`internal/api/`)
   - RESTful endpoints for proxy lifecycle management
   - Middleware for logging and recovery
   - Routes defined in handler.go:25-40

### Critical Architecture Details

#### Mihomo Multi-Port Implementation

The key innovation is using Mihomo's `WithSpecialProxy` option to create isolated listeners:

```go
// Each proxy instance gets its own port and node binding
socks.New(addr, tunnel, inbound.WithSpecialProxy(proxy.Name()))
```

This is implemented in `internal/mihomo/adapter.go:157-158`. Each listener is completely independent and routes traffic through its designated node only.

#### Proxy Lifecycle

1. **Creation** (Manager.CreateProxy in proxy/manager.go:71):
   - Select node based on filters (availability, delay, type, region)
   - Allocate available port from pool
   - Create Mihomo listener bound to specific node
   - Register instance with TTL

2. **Active State**:
   - Client connects to allocated port (e.g., 127.0.0.1:20001)
   - Traffic routes through bound node automatically
   - Instance tracked in Manager.instances map

3. **Cleanup** (TTLCleaner in proxy/ttl_cleaner.go):
   - Scans every 30s for expired instances
   - Calls Manager.ReleaseProxy to close listener and free port
   - Removes from tracking maps

#### Node Selection Algorithm

Located in `proxy/manager.go:selectNode()`:
- Filters nodes by: available=true, maxDelay, nodeType, region
- Sorts by delay (lowest first)
- Returns best match or nil if no nodes available

#### Health Checking

Background goroutine in `node/health_checker.go`:
- Tests each node periodically (default: 5m interval)
- Makes HTTP request to test URL (default: google.com/generate_204)
- Updates node.Available and node.Delay fields
- Timeout: 10s, MaxDelay threshold: 1000ms

## Configuration

Edit `configs/config.yaml` before starting:

### Critical Settings

```yaml
subscription:
  sources:
    - name: "主订阅"
      url: "YOUR_SUBSCRIPTION_URL"
      enabled: true  # Must be true to fetch nodes
  update_interval: 6h  # How often to refresh subscriptions

proxy:
  port_range:
    min: 20000  # Start of dynamic port range
    max: 30000  # End of dynamic port range
  default_ttl: 30m  # Default proxy instance lifetime
  max_instances: 100  # Maximum concurrent proxies

health_check:
  enabled: true  # Disable if nodes are pre-validated
  interval: 5m
  max_delay: 1000  # Nodes slower than this marked unavailable
```

## API Endpoints

### Proxy Management
- `POST /api/getProxy` - Create new proxy instance (returns host:port)
- `POST /api/releaseProxy` - Manually release proxy before TTL expires
- `GET /api/getInstance/:id` - Get specific instance details
- `GET /api/listInstances` - List all active instances
- `GET /api/stats` - Get system statistics

### Node Management
- `GET /api/nodePool` - View node pool status (total, available, by type)
- `GET /api/nodes` - List all nodes with health status
- `POST /api/subscription/update` - Manually trigger subscription update

### Health
- `GET /health` - Service health check

## Important Implementation Notes

### Mihomo Integration Constraints

- **Mixed protocol not supported**: Use `socks5` or `http` explicitly (see adapter.go:178-181)
- **Global tunnel instance**: Mihomo uses singleton pattern, all listeners share tunnel.Tunnel
- **Proxy registration required**: Each node must be registered in proxyRegistry before creating listener (adapter.go:116-127)

### Concurrency Patterns

- All managers use `sync.RWMutex` for thread-safety
- Subscription updates run in parallel goroutines per source (subscription/manager.go:86-108)
- TTL cleaner runs in background goroutine, scans periodically
- Health checker runs in background goroutine per node

### Port Management

- Ports allocated sequentially from min to max
- Released ports immediately available for reuse
- Port conflicts prevented by PortAllocator.allocated map
- If port range exhausted, getProxy returns error

### Error Handling

- Node selection returns nil if no nodes match filters
- Port allocation fails if range exhausted
- Mihomo listener creation can fail (port in use, invalid node config)
- All errors propagate up to API layer with descriptive messages

## Common Development Patterns

### Adding New Node Filters

Modify `node/pool.go:Filter()` to add new FilterOptions fields and filtering logic.

### Adding New Protocols

1. Add case in `mihomo/adapter.go:CreateListener()` switch statement
2. Import appropriate Mihomo listener package
3. Call listener constructor with `WithSpecialProxy` option
4. Update CloseListener to handle new listener type

### Modifying TTL Behavior

- Change scan interval: `proxy/ttl_cleaner.go` constructor parameter
- Change default TTL: `configs/config.yaml` proxy.default_ttl
- Per-request TTL: Pass `ttl` field in POST /api/getProxy body

## File Structure Reference

```
cmd/server/main.go              # Application entry point, initialization sequence
internal/
├── config/                     # YAML configuration loading
├── subscription/               # Subscription fetching and parsing
│   └── manager.go             # Auto-update orchestration
├── node/                       # Node pool and health checking
│   ├── pool.go                # Thread-safe node registry
│   └── health_checker.go      # Background health monitoring
├── proxy/                      # Proxy lifecycle management
│   ├── manager.go             # Core orchestrator (CreateProxy/ReleaseProxy)
│   ├── port_allocator.go      # Dynamic port allocation
│   └── ttl_cleaner.go         # Background TTL enforcement
├── mihomo/                     # Mihomo core integration
│   └── adapter.go             # Multi-port listener management
└── api/                        # HTTP API layer
    └── handler.go             # Route definitions and handlers
```