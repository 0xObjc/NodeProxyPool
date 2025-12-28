package main

import (
	"fmt"
	"time"

	"proxyPool/internal/mihomo"
	"proxyPool/pkg/logger"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/outbound"
)

func main() {
	// 初始化日志
	if err := logger.Init("info", ""); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		return
	}
	defer logger.Sync()

	fmt.Println("=== Mihomo Multi-Port Test ===")

	// 创建Mihomo适配器
	mihomoAdapter := mihomo.NewMihomoAdapter()

	// 初始化
	fmt.Println("\n1. Initializing Mihomo...")
	if err := mihomoAdapter.Init(); err != nil {
		fmt.Printf("Failed to initialize: %v\n", err)
		return
	}
	fmt.Println("✓ Mihomo initialized")

	// 创建Direct代理(用于测试)
	directOutbound := outbound.NewDirect()
	directProxy := adapter.NewProxy(directOutbound)

	// 测试创建单个SOCKS5代理
	fmt.Println("\n2. Creating SOCKS5 listener on port 20001...")
	if err := mihomoAdapter.CreateListener("socks5", 20001, directProxy); err != nil {
		fmt.Printf("Failed to create listener: %v\n", err)
		return
	}
	fmt.Println("✓ SOCKS5 listener created on port 20001")

	// 测试创建第二个SOCKS5代理(不同端口)
	fmt.Println("\n3. Creating SOCKS5 listener on port 20002...")
	if err := mihomoAdapter.CreateListener("socks5", 20002, directProxy); err != nil {
		fmt.Printf("Failed to create listener: %v\n", err)
		return
	}
	fmt.Println("✓ SOCKS5 listener created on port 20002")

	// 测试创建HTTP代理
	fmt.Println("\n4. Creating HTTP listener on port 20003...")
	if err := mihomoAdapter.CreateListener("http", 20003, directProxy); err != nil {
		fmt.Printf("Failed to create listener: %v\n", err)
		return
	}
	fmt.Println("✓ HTTP listener created on port 20003")

	fmt.Printf("\n✓ Successfully created %d listeners!\n", mihomoAdapter.ListenerCount())

	// 测试端口冲突检测
	fmt.Println("\n5. Testing port conflict detection...")
	if err := mihomoAdapter.CreateListener("socks5", 20001, directProxy); err != nil {
		fmt.Printf("✓ Port conflict correctly detected: %v\n", err)
	} else {
		fmt.Println("✗ Port conflict not detected!")
	}

	// 保持运行一段时间以便手动测试
	fmt.Println("\n6. Listeners are running. Test with curl:")
	fmt.Println("   curl -x socks5://127.0.0.1:20001 https://api.ipify.org")
	fmt.Println("   curl -x socks5://127.0.0.1:20002 https://api.ipify.org")
	fmt.Println("   curl -x http://127.0.0.1:20003 https://api.ipify.org")
	fmt.Println("\nPress Ctrl+C to stop...")

	time.Sleep(30 * time.Second)

	// 清理
	fmt.Println("\n7. Closing listeners...")
	mihomoAdapter.CloseListener(20001)
	mihomoAdapter.CloseListener(20002)
	mihomoAdapter.CloseListener(20003)
	fmt.Println("✓ All listeners closed")

	fmt.Println("\n=== Test Complete ===")
}
