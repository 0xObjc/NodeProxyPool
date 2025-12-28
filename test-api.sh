#!/bin/bash

# ProxyPool API 测试脚本

BASE_URL="http://localhost:8080"

echo "========================================="
echo "ProxyPool API 测试脚本"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4

    echo -e "${YELLOW}测试: ${name}${NC}"
    echo "请求: ${method} ${endpoint}"

    if [ -z "$data" ]; then
        response=$(curl -s -X ${method} "${BASE_URL}${endpoint}")
    else
        echo "数据: ${data}"
        response=$(curl -s -X ${method} "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            -d "${data}")
    fi

    echo "响应: ${response}"
    echo ""
}

# 1. 健康检查
echo "========================================="
echo "1. 健康检查"
echo "========================================="
test_api "健康检查" "GET" "/health"

# 2. 节点池状态
echo "========================================="
echo "2. 节点池状态"
echo "========================================="
test_api "获取节点池状态" "GET" "/api/nodePool"

# 3. 节点列表
echo "========================================="
echo "3. 节点列表"
echo "========================================="
test_api "获取节点列表" "GET" "/api/nodes"

# 4. 获取代理
echo "========================================="
echo "4. 获取代理"
echo "========================================="
test_api "获取代理(默认参数)" "POST" "/api/getProxy" '{}'

# 5. 统计信息
echo "========================================="
echo "5. 统计信息"
echo "========================================="
test_api "获取统计信息" "GET" "/api/stats"

# 6. 列出所有实例
echo "========================================="
echo "6. 列出所有实例"
echo "========================================="
test_api "列出所有实例" "GET" "/api/listInstances"

# 7. 手动更新订阅
echo "========================================="
echo "7. 手动更新订阅"
echo "========================================="
test_api "手动更新订阅" "POST" "/api/subscription/update"

echo "========================================="
echo "测试完成"
echo "========================================="
