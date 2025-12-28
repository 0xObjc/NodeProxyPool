package test_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type respModel struct {
	Code int `json:"code"`
	Data struct {
		CreatedAt  time.Time `json:"created_at"`
		ExpiresAt  time.Time `json:"expires_at"`
		Host       string    `json:"host"`
		InstanceId string    `json:"instance_id"`
		NodeDelay  int       `json:"node_delay"`
		NodeName   string    `json:"node_name"`
		Port       int       `json:"port"`
		Protocol   string    `json:"protocol"`
		Remaining  int       `json:"remaining"`
		Ttl        int       `json:"ttl"`
	} `json:"data"`
	Message string `json:"message"`
}

func TestGetProxy(t *testing.T) {
	httpClient := &http.Client{}

	body := bytes.NewBuffer([]byte(`{}`))
	req, err := http.NewRequest("POST", "http://localhost:8080/api/getProxy", body)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	respModel := respModel{}
	err = json.Unmarshal(content, &respModel)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(respModel)

	// 使用这个proxy请求ipinfo.io

	proxyUrl := fmt.Sprintf("%s://%s:%d", respModel.Data.Protocol, respModel.Data.Host, respModel.Data.Port)
	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		t.Fatal(err)
	}

	httpClient.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	req, err = http.NewRequest("GET", "http://ipinfo.io", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	content, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(content))
}
