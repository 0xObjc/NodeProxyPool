package subscription

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher 订阅抓取器
type Fetcher struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewFetcher 创建抓取器
func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Fetch 抓取订阅内容
func (f *Fetcher) Fetch(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置User-Agent,模拟常见客户端
	req.Header.Set("User-Agent", "clash-verge/v2.4.3")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return content, nil
}
