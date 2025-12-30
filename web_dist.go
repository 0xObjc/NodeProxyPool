package proxypool

import "embed"

// WebDist 导出前端构建产物
//go:embed all:web/dist
var WebDist embed.FS
