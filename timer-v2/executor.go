package timer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// Exec 根据任务类型分发到对应的执行器
func Exec(t *Timer) (interface{}, error) {
	switch t.Type {
	case "http":
		return execHTTP(t.Content)
	case "shell":
		return execShell(t.Content)
	case "webhook":
		return execWebhook(t.Content)
	default: // "script" 或空值
		return Script.Exec(t.Content)
	}
}

// HTTPConfig HTTP 请求配置
type HTTPConfig struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// execHTTP 发送 HTTP 请求
func execHTTP(content string) (interface{}, error) {
	var cfg HTTPConfig
	if err := json.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("解析HTTP配置失败: %w", err)
	}
	if cfg.Method == "" {
		cfg.Method = "GET"
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("URL不能为空")
	}

	var bodyReader io.Reader
	if cfg.Body != "" {
		bodyReader = bytes.NewReader([]byte(cfg.Body))
	}

	req, err := http.NewRequest(cfg.Method, cfg.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

// execShell 执行系统命令
func execShell(content string) (interface{}, error) {
	if content == "" {
		return nil, fmt.Errorf("命令不能为空")
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", content)
	} else {
		cmd = exec.Command("sh", "-c", content)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		return string(output), fmt.Errorf("执行失败: %w\n%s", err, string(output))
	} else {
		return string(output), nil
	}
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	URL     string `json:"url"`
	Payload string `json:"payload"`
}

// execWebhook 发送 Webhook 通知
func execWebhook(content string) (interface{}, error) {
	var cfg WebhookConfig
	if err := json.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("解析Webhook配置失败: %w", err)
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("URL不能为空")
	}

	resp, err := http.Post(cfg.URL, "application/json", bytes.NewReader([]byte(cfg.Payload)))
	if err != nil {
		return nil, fmt.Errorf("Webhook发送失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Webhook HTTP %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}
