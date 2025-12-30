package subscription

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"proxyPool/internal/node"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/common/convert"
	"github.com/metacubex/mihomo/config"
	"gopkg.in/yaml.v3"
)

// Parser 节点解析器
type Parser struct{}

// NewParser 创建解析器
func NewParser() *Parser {
	return &Parser{}
}

// Parse 解析订阅内容为节点列表
func (p *Parser) Parse(content []byte) ([]*node.Node, error) {
	decoded := decodePermissive(content)

	// 优先检测 YAML (Clash/Mihomo) 配置
	if p.isLikelyYAML(decoded) {
		if nodes, ok := p.tryParseYAMLProxies(decoded); ok {
			return nodes, nil
		}
		if nodes, ok := p.tryParseConfig(decoded); ok {
			return nodes, nil
		}
	}

	// 使用Mihomo的ConvertsV2Ray解析节点
	proxies, err := convert.ConvertsV2Ray(decoded)
	if err != nil {
		// 兜底: 尝试按 Clash/Mihomo YAML 订阅解析
		if nodes, ok := p.tryParseConfig(decoded); ok {
			return nodes, nil
		}
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}

	nodes := make([]*node.Node, 0, len(proxies))
	for _, proxyConfig := range proxies {
		// 创建节点对象
		n := &node.Node{
			ID:        generateNodeID(proxyConfig),
			RawConfig: proxyConfig,
			Available: true, // 初始标记为可用
			Delay:     -1,   // 待健康检查
		}

		// 提取基本信息
		if name, ok := proxyConfig["name"].(string); ok {
			n.Name = name
		}
		if proxyType, ok := proxyConfig["type"].(string); ok {
			n.Type = proxyType
		}
		if server, ok := proxyConfig["server"].(string); ok {
			n.Server = server
		}
		if port, ok := proxyConfig["port"].(int); ok {
			n.Port = port
		}

		// 创建Mihomo Proxy对象
		proxy, err := adapter.ParseProxy(proxyConfig)
		if err != nil {
			// 如果解析失败,跳过此节点
			continue
		}
		n.Proxy = proxy

		nodes = append(nodes, n)
	}

	return nodes, nil
}

// generateNodeID 生成节点唯一ID
func generateNodeID(config map[string]interface{}) string {
	// 使用name+server+port生成MD5作为ID
	name, _ := config["name"].(string)
	server, _ := config["server"].(string)
	port, _ := config["port"].(int)

	data := fmt.Sprintf("%s-%s-%d", name, server, port)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// decodePermissive 尝试宽松的Base64解码(支持缺少填充/URL变体),失败则返回原文
func decodePermissive(content []byte) []byte {
	raw := strings.TrimSpace(string(content))
	if raw == "" {
		return content
	}

	decoders := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	for _, enc := range decoders {
		if decoded, err := enc.DecodeString(raw); err == nil {
			return decoded
		}
	}
	return []byte(raw)
}

// tryParseConfig 兜底: 解析 Clash/Mihomo YAML 配置里的 proxies 段
func (p *Parser) tryParseConfig(buf []byte) ([]*node.Node, bool) {
	cfg, err := config.Parse(buf)
	if err != nil || len(cfg.Proxies) == 0 {
		return nil, false
	}

	nodes := make([]*node.Node, 0, len(cfg.Proxies))
	for name, proxy := range cfg.Proxies {
		n := &node.Node{
			ID:        fmt.Sprintf("%x", md5.Sum([]byte(name))),
			Name:      name,
			Type:      strings.ToLower(proxy.Type().String()),
			Proxy:     proxy,
			Available: true,
			Delay:     -1,
		}
		nodes = append(nodes, n)
	}
	return nodes, true
}

// tryParseYAMLProxies 只读取 YAML 中的 proxies 段, 忽略其他字段
func (p *Parser) tryParseYAMLProxies(buf []byte) ([]*node.Node, bool) {
	var wrapper struct {
		Proxies []map[string]interface{} `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(buf, &wrapper); err != nil || len(wrapper.Proxies) == 0 {
		return nil, false
	}

	nodes := make([]*node.Node, 0, len(wrapper.Proxies))
	for _, proxyConfig := range wrapper.Proxies {
		normalizeProxyConfig(proxyConfig)

		n := &node.Node{
			ID:        generateNodeID(proxyConfig),
			RawConfig: proxyConfig,
			Available: true,
			Delay:     -1,
		}
		if name, ok := proxyConfig["name"].(string); ok {
			n.Name = name
		}
		if proxyType, ok := proxyConfig["type"].(string); ok {
			n.Type = proxyType
		}
		if server, ok := proxyConfig["server"].(string); ok {
			n.Server = server
		}
		if port, ok := proxyConfig["port"].(int); ok {
			n.Port = port
		}

		proxy, err := adapter.ParseProxy(proxyConfig)
		if err != nil {
			continue
		}
		n.Proxy = proxy
		nodes = append(nodes, n)
	}

	if len(nodes) == 0 {
		return nil, false
	}
	return nodes, true
}

// isLikelyYAML 粗略判断内容是否为 Clash/Mihomo YAML
func (p *Parser) isLikelyYAML(buf []byte) bool {
	raw := strings.TrimSpace(string(buf))
	if raw == "" {
		return false
	}
	indicators := []string{
		"proxies:",
		"proxy-groups:",
		"mixed-port:",
		"mode:",
		"rules:",
	}
	for _, k := range indicators {
		if strings.Contains(raw, k) {
			return true
		}
	}
	return false
}

// normalizeProxyConfig 尝试把通用 YAML 里字符串/切片等转成适合 ParseProxy 的格式
func normalizeProxyConfig(cfg map[string]interface{}) {
	// 如果 password 是切片, 取第一个元素
	if pwdSlice, ok := cfg["password"].([]interface{}); ok && len(pwdSlice) > 0 {
		if s, ok := pwdSlice[0].(string); ok {
			cfg["password"] = s
		}
	}

	// 端口转为 int
	switch v := cfg["port"].(type) {
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			cfg["port"] = n
		}
	}
}
