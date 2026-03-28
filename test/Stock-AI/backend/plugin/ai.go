package plugin

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AIChat 使用AI插件进行对话
func (m *Manager) AIChat(pluginID string, messages []AIChatMessage) (string, error) {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return "", err
	}

	if plugin.Type != PluginTypeAI {
		return "", fmt.Errorf("插件类型错误: %s", plugin.Type)
	}

	if !plugin.Enabled {
		return "", fmt.Errorf("插件未启用: %s", plugin.Name)
	}

	var config AIConfig
	if err := json.Unmarshal(plugin.Config, &config); err != nil {
		return "", fmt.Errorf("解析AI配置失败: %w", err)
	}

	return m.executeAIChat(&config, messages, false)
}

// AIChatStream 使用AI插件进行流式对话
func (m *Manager) AIChatStream(pluginID string, messages []AIChatMessage) (<-chan string, error) {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	if plugin.Type != PluginTypeAI {
		return nil, fmt.Errorf("插件类型错误: %s", plugin.Type)
	}

	if !plugin.Enabled {
		return nil, fmt.Errorf("插件未启用: %s", plugin.Name)
	}

	var config AIConfig
	if err := json.Unmarshal(plugin.Config, &config); err != nil {
		return nil, fmt.Errorf("解析AI配置失败: %w", err)
	}

	return m.executeAIChatStream(&config, messages)
}

// AIChatFromAll 从所有启用的AI插件中选择一个进行对话
func (m *Manager) AIChatFromAll(messages []AIChatMessage) (string, string, error) {
	plugins := m.GetPluginsByType(PluginTypeAI)

	for _, plugin := range plugins {
		if !plugin.Enabled {
			continue
		}

		result, err := m.AIChat(plugin.ID, messages)
		if err == nil && result != "" {
			return result, plugin.ID, nil
		}
	}

	return "", "", fmt.Errorf("所有AI插件都无法响应")
}

// TestAI 测试AI插件
func (m *Manager) TestAI(pluginID string) (string, error) {
	testMessages := []AIChatMessage{
		{Role: "user", Content: "你好，请简单介绍一下你自己。"},
	}
	return m.AIChat(pluginID, testMessages)
}

// GetEnabledAIPlugins 获取所有启用的AI插件
func (m *Manager) GetEnabledAIPlugins() []Plugin {
	plugins := m.GetPluginsByType(PluginTypeAI)
	var enabled []Plugin
	for _, p := range plugins {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled
}

// HasEnabledAIPlugins 检查是否有启用的AI插件
func (m *Manager) HasEnabledAIPlugins() bool {
	return len(m.GetEnabledAIPlugins()) > 0
}

// executeAIChat 执行AI对话
func (m *Manager) executeAIChat(config *AIConfig, messages []AIChatMessage, stream bool) (string, error) {
	// 构建请求URL
	baseURL := strings.TrimSuffix(config.BaseURL, "/")
	url := baseURL + "/chat/completions"

	// 构建请求体
	reqMessages := messages
	if config.SystemPrompt != "" {
		// 在消息列表前添加系统提示
		reqMessages = append([]AIChatMessage{{Role: "system", Content: config.SystemPrompt}}, messages...)
	}

	reqBody := AIChatRequest{
		Model:       config.Model,
		Messages:    reqMessages,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		Stream:      stream,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("请求失败: %d - %s", resp.StatusCode, string(body))
	}

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var respData AIChatResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(respData.Choices) == 0 {
		return "", fmt.Errorf("AI未返回有效响应")
	}

	return respData.Choices[0].Message.Content, nil
}

// executeAIChatStream 执行流式AI对话
func (m *Manager) executeAIChatStream(config *AIConfig, messages []AIChatMessage) (<-chan string, error) {
	// 构建请求URL
	baseURL := strings.TrimSuffix(config.BaseURL, "/")
	url := baseURL + "/chat/completions"

	// 构建请求体
	reqMessages := messages
	if config.SystemPrompt != "" {
		reqMessages = append([]AIChatMessage{{Role: "system", Content: config.SystemPrompt}}, messages...)
	}

	reqBody := AIChatRequest{
		Model:       config.Model,
		Messages:    reqMessages,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		Stream:      true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("请求失败: %d - %s", resp.StatusCode, string(body))
	}

	// 创建输出通道
	ch := make(chan string, 100)

	// 启动goroutine处理流式响应
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					ch <- fmt.Sprintf("\n[错误: %v]", err)
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// SSE格式: data: {...}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			// 解析JSON
			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				continue
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				ch <- streamResp.Choices[0].Delta.Content
			}
		}
	}()

	return ch, nil
}

// 预置AI模型模板
var AITemplates = []struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Config      AIConfig `json:"config"`
}{
	{
		ID:          "deepseek",
		Name:        "DeepSeek",
		Description: "DeepSeek AI模型，性价比高",
		Config: AIConfig{
			Provider:     "openai-compatible",
			BaseURL:      "https://api.deepseek.com/v1",
			APIKey:       "",
			Model:        "deepseek-chat",
			MaxTokens:    4096,
			Temperature:  0.7,
			SystemPrompt: "你是一个专业的股票分析助手。",
		},
	},
	{
		ID:          "qwen",
		Name:        "通义千问",
		Description: "阿里云通义千问大模型",
		Config: AIConfig{
			Provider:     "openai-compatible",
			BaseURL:      "https://dashscope.aliyuncs.com/compatible-mode/v1",
			APIKey:       "",
			Model:        "qwen-turbo",
			MaxTokens:    4096,
			Temperature:  0.7,
			SystemPrompt: "你是一个专业的股票分析助手。",
		},
	},
	{
		ID:          "zhipu",
		Name:        "智谱GLM",
		Description: "清华智谱GLM大模型",
		Config: AIConfig{
			Provider:     "openai-compatible",
			BaseURL:      "https://open.bigmodel.cn/paas/v4",
			APIKey:       "",
			Model:        "glm-4-flash",
			MaxTokens:    4096,
			Temperature:  0.7,
			SystemPrompt: "你是一个专业的股票分析助手。",
		},
	},
	{
		ID:          "siliconflow",
		Name:        "硅基流动",
		Description: "硅基流动聚合多种AI模型",
		Config: AIConfig{
			Provider:     "openai-compatible",
			BaseURL:      "https://api.siliconflow.cn/v1",
			APIKey:       "",
			Model:        "Qwen/Qwen2.5-7B-Instruct",
			MaxTokens:    4096,
			Temperature:  0.7,
			SystemPrompt: "你是一个专业的股票分析助手。",
		},
	},
	{
		ID:          "ollama",
		Name:        "Ollama本地模型",
		Description: "本地部署的Ollama模型",
		Config: AIConfig{
			Provider:     "openai-compatible",
			BaseURL:      "http://localhost:11434/v1",
			APIKey:       "",
			Model:        "qwen3.5:397b-cloud",
			MaxTokens:    4096,
			Temperature:  0.7,
			SystemPrompt: "你是一个专业的股票分析助手。",
		},
	},
	{
		ID:          "custom-openai",
		Name:        "自定义OpenAI兼容接口",
		Description: "接入任何OpenAI兼容的API",
		Config: AIConfig{
			Provider:     "openai-compatible",
			BaseURL:      "",
			APIKey:       "",
			Model:        "",
			MaxTokens:    4096,
			Temperature:  0.7,
			SystemPrompt: "你是一个专业的股票分析助手。",
		},
	},
}

// GetAITemplates 获取预置AI模板
func (m *Manager) GetAITemplates() []struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Config      AIConfig `json:"config"`
} {
	return AITemplates
}

// CreateAIPluginFromTemplate 从模板创建AI插件
func (m *Manager) CreateAIPluginFromTemplate(templateID string, name string, apiKey string, baseURL string, model string) (*Plugin, error) {
	var template *struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Config      AIConfig `json:"config"`
	}

	for i := range AITemplates {
		if AITemplates[i].ID == templateID {
			template = &AITemplates[i]
			break
		}
	}

	if template == nil {
		return nil, fmt.Errorf("模板不存在: %s", templateID)
	}

	// 复制配置
	config := template.Config
	config.APIKey = apiKey
	if baseURL != "" {
		config.BaseURL = baseURL
	}
	if model != "" {
		config.Model = model
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}

	plugin := &Plugin{
		ID:          fmt.Sprintf("%s-%d", templateID, m.generateID()),
		Name:        name,
		Type:        PluginTypeAI,
		Description: template.Description,
		Enabled:     true,
		Config:      configJSON,
	}

	return plugin, nil
}

// generateID 生成唯一ID
func (m *Manager) generateID() int64 {
	return int64(len(m.plugins)) + 1
}
