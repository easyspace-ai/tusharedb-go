package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Manager 插件管理器
type Manager struct {
	pluginsDir string
	plugins    map[string]*Plugin
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewManager 创建插件管理器
func NewManager(pluginsDir string) *Manager {
	return &Manager{
		pluginsDir: pluginsDir,
		plugins:    make(map[string]*Plugin),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Init 初始化插件管理器
func (m *Manager) Init() error {
	// 确保插件目录存在
	if err := os.MkdirAll(m.pluginsDir, 0755); err != nil {
		return fmt.Errorf("创建插件目录失败: %w", err)
	}

	// 加载所有插件
	return m.LoadPlugins()
}

// LoadPlugins 加载所有插件
func (m *Manager) LoadPlugins() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	configPath := filepath.Join(m.pluginsDir, "plugins.json")

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := []Plugin{}
		data, _ := json.MarshalIndent(defaultConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("创建默认插件配置失败: %w", err)
		}
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取插件配置失败: %w", err)
	}

	var plugins []Plugin
	if err := json.Unmarshal(data, &plugins); err != nil {
		return fmt.Errorf("解析插件配置失败: %w", err)
	}

	// 加载到内存
	m.plugins = make(map[string]*Plugin)
	for i := range plugins {
		m.plugins[plugins[i].ID] = &plugins[i]
	}

	return nil
}

// SavePlugins 保存所有插件配置
func (m *Manager) SavePlugins() error {
	m.mu.RLock()
	plugins := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, *p)
	}
	m.mu.RUnlock()

	data, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化插件配置失败: %w", err)
	}

	configPath := filepath.Join(m.pluginsDir, "plugins.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存插件配置失败: %w", err)
	}

	return nil
}

// GetPlugins 获取所有插件
func (m *Manager) GetPlugins() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]Plugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		plugins = append(plugins, *p)
	}
	return plugins
}

// GetPluginsByType 按类型获取插件
func (m *Manager) GetPluginsByType(pluginType PluginType) []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var plugins []Plugin
	for _, p := range m.plugins {
		if p.Type == pluginType {
			plugins = append(plugins, *p)
		}
	}
	return plugins
}

// GetPlugin 获取单个插件
func (m *Manager) GetPlugin(id string) (*Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, ok := m.plugins[id]
	if !ok {
		return nil, fmt.Errorf("插件不存在: %s", id)
	}
	return plugin, nil
}

// AddPlugin 添加插件
func (m *Manager) AddPlugin(plugin *Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[plugin.ID]; exists {
		return fmt.Errorf("插件ID已存在: %s", plugin.ID)
	}

	plugin.CreatedAt = time.Now()
	plugin.UpdatedAt = time.Now()
	m.plugins[plugin.ID] = plugin

	// 保存到文件
	m.mu.Unlock()
	err := m.SavePlugins()
	m.mu.Lock()
	return err
}

// UpdatePlugin 更新插件
func (m *Manager) UpdatePlugin(plugin *Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[plugin.ID]; !exists {
		return fmt.Errorf("插件不存在: %s", plugin.ID)
	}

	plugin.UpdatedAt = time.Now()
	m.plugins[plugin.ID] = plugin

	// 保存到文件
	m.mu.Unlock()
	err := m.SavePlugins()
	m.mu.Lock()
	return err
}

// DeletePlugin 删除插件
func (m *Manager) DeletePlugin(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[id]; !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	delete(m.plugins, id)

	// 保存到文件
	m.mu.Unlock()
	err := m.SavePlugins()
	m.mu.Lock()
	return err
}

// TogglePlugin 启用/禁用插件
func (m *Manager) TogglePlugin(id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("插件不存在: %s", id)
	}

	plugin.Enabled = enabled
	plugin.UpdatedAt = time.Now()

	// 保存到文件
	m.mu.Unlock()
	err := m.SavePlugins()
	m.mu.Lock()
	return err
}

// SendNotification 发送通知
func (m *Manager) SendNotification(pluginID string, data *NotificationData) error {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return err
	}

	if plugin.Type != PluginTypeNotification {
		return fmt.Errorf("插件类型错误: %s", plugin.Type)
	}

	if !plugin.Enabled {
		return fmt.Errorf("插件未启用: %s", plugin.Name)
	}

	var config NotificationConfig
	if err := json.Unmarshal(plugin.Config, &config); err != nil {
		return fmt.Errorf("解析通知配置失败: %w", err)
	}

	return m.executeNotification(&config, data)
}

// SendNotificationToAll 发送通知到所有启用的通知插件
func (m *Manager) SendNotificationToAll(data *NotificationData) []error {
	plugins := m.GetPluginsByType(PluginTypeNotification)
	var errors []error

	for _, plugin := range plugins {
		if !plugin.Enabled {
			continue
		}

		if err := m.SendNotification(plugin.ID, data); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", plugin.Name, err))
		}
	}

	return errors
}

// TestNotification 测试通知
func (m *Manager) TestNotification(pluginID string) error {
	testData := &NotificationData{
		StockCode:     "000001",
		StockName:     "平安银行",
		AlertType:     "股价提醒",
		CurrentPrice:  10.50,
		Condition:     "高于",
		TargetValue:   10.00,
		TriggerTime:   time.Now().Format("2006-01-02 15:04:05"),
		Change:        0.25,
		ChangePercent: 2.44,
	}

	return m.SendNotification(pluginID, testData)
}

// executeNotification 执行通知发送
func (m *Manager) executeNotification(config *NotificationConfig, data *NotificationData) error {
	// 替换URL中的变量
	reqURL := m.replaceVariables(config.URL, data, config.Params)

	var reqBody io.Reader
	var contentType string

	switch config.ContentType {
	case "json":
		// JSON格式
		bodyStr := m.replaceVariablesInJSON(config.BodyTemplate, data)
		reqBody = strings.NewReader(bodyStr)
		contentType = "application/json"
	case "form":
		// 表单格式
		bodyStr := m.replaceVariables(config.BodyTemplate.(string), data, config.Params)
		reqBody = strings.NewReader(bodyStr)
		contentType = "application/x-www-form-urlencoded"
	case "text":
		// 纯文本
		if bodyTemplate, ok := config.BodyTemplate.(string); ok {
			bodyStr := m.replaceVariables(bodyTemplate, data, config.Params)
			reqBody = strings.NewReader(bodyStr)
		}
		contentType = "text/plain"
	default:
		// 默认JSON
		bodyStr := m.replaceVariablesInJSON(config.BodyTemplate, data)
		reqBody = strings.NewReader(bodyStr)
		contentType = "application/json"
	}

	req, err := http.NewRequest(config.Method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range config.Headers {
		req.Header.Set(key, m.replaceVariables(value, data, config.Params))
	}
	if req.Header.Get("Content-Type") == "" && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// 发送请求
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// replaceVariables 替换字符串中的变量
func (m *Manager) replaceVariables(template string, data *NotificationData, params map[string]string) string {
	result := template

	// 替换通知数据变量
	result = strings.ReplaceAll(result, "{stockCode}", data.StockCode)
	result = strings.ReplaceAll(result, "{stockName}", data.StockName)
	result = strings.ReplaceAll(result, "{alertType}", data.AlertType)
	result = strings.ReplaceAll(result, "{currentPrice}", fmt.Sprintf("%.2f", data.CurrentPrice))
	result = strings.ReplaceAll(result, "{condition}", data.Condition)
	result = strings.ReplaceAll(result, "{targetValue}", fmt.Sprintf("%.2f", data.TargetValue))
	result = strings.ReplaceAll(result, "{triggerTime}", data.TriggerTime)
	result = strings.ReplaceAll(result, "{change}", fmt.Sprintf("%.2f", data.Change))
	result = strings.ReplaceAll(result, "{changePercent}", fmt.Sprintf("%.2f%%", data.ChangePercent))

	// 替换自定义参数
	for key, value := range params {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}

	// URL编码（如果是URL的一部分）
	if strings.Contains(template, "http") {
		// 对特殊字符进行URL编码
		re := regexp.MustCompile(`/([^/]+)$`)
		result = re.ReplaceAllStringFunc(result, func(s string) string {
			parts := strings.SplitN(s, "/", 2)
			if len(parts) == 2 {
				return "/" + url.PathEscape(parts[1])
			}
			return s
		})
	}

	return result
}

// replaceVariablesInJSON 替换JSON中的变量
func (m *Manager) replaceVariablesInJSON(template interface{}, data *NotificationData) string {
	jsonBytes, err := json.Marshal(template)
	if err != nil {
		return "{}"
	}

	result := string(jsonBytes)

	// 替换通知数据变量
	result = strings.ReplaceAll(result, "{stockCode}", data.StockCode)
	result = strings.ReplaceAll(result, "{stockName}", data.StockName)
	result = strings.ReplaceAll(result, "{alertType}", data.AlertType)
	result = strings.ReplaceAll(result, "{currentPrice}", fmt.Sprintf("%.2f", data.CurrentPrice))
	result = strings.ReplaceAll(result, "{condition}", data.Condition)
	result = strings.ReplaceAll(result, "{targetValue}", fmt.Sprintf("%.2f", data.TargetValue))
	result = strings.ReplaceAll(result, "{triggerTime}", data.TriggerTime)
	result = strings.ReplaceAll(result, "{change}", fmt.Sprintf("%.2f", data.Change))
	result = strings.ReplaceAll(result, "{changePercent}", fmt.Sprintf("%.2f%%", data.ChangePercent))

	return result
}

// GetNotificationTemplates 获取预置通知模板
func (m *Manager) GetNotificationTemplates() []NotificationTemplate {
	return NotificationTemplates
}

// CreatePluginFromTemplate 从模板创建插件
func (m *Manager) CreatePluginFromTemplate(templateID string, name string, params map[string]string) (*Plugin, error) {
	var template *NotificationTemplate
	for _, t := range NotificationTemplates {
		if t.ID == templateID {
			template = &t
			break
		}
	}

	if template == nil {
		return nil, fmt.Errorf("模板不存在: %s", templateID)
	}

	// 复制配置并填入参数
	config := template.Config
	for key, value := range params {
		config.Params[key] = value
	}

	// 如果URL中有参数占位符，替换它
	for key, value := range params {
		config.URL = strings.ReplaceAll(config.URL, "{"+key+"}", value)
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("序列化配置失败: %w", err)
	}

	plugin := &Plugin{
		ID:          fmt.Sprintf("%s-%d", templateID, time.Now().UnixNano()),
		Name:        name,
		Type:        PluginTypeNotification,
		Description: template.Description,
		Enabled:     true,
		Config:      configJSON,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return plugin, nil
}

// ValidateNotificationConfig 验证通知配置
func (m *Manager) ValidateNotificationConfig(config *NotificationConfig) error {
	if config.URL == "" {
		return fmt.Errorf("URL不能为空")
	}

	if config.Method == "" {
		config.Method = "POST"
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true}
	if !validMethods[config.Method] {
		return fmt.Errorf("不支持的HTTP方法: %s", config.Method)
	}

	return nil
}

// GetEnabledNotificationPlugins 获取所有启用的通知插件
func (m *Manager) GetEnabledNotificationPlugins() []Plugin {
	plugins := m.GetPluginsByType(PluginTypeNotification)
	var enabled []Plugin
	for _, p := range plugins {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled
}

// HasEnabledNotificationPlugins 检查是否有启用的通知插件
func (m *Manager) HasEnabledNotificationPlugins() bool {
	return len(m.GetEnabledNotificationPlugins()) > 0
}

// ImportPlugin 从JSON字符串导入插件
func (m *Manager) ImportPlugin(jsonData string) (*Plugin, error) {
	var plugin Plugin
	if err := json.Unmarshal([]byte(jsonData), &plugin); err != nil {
		return nil, fmt.Errorf("解析插件JSON失败: %w", err)
	}

	// 验证必填字段
	if plugin.ID == "" {
		return nil, fmt.Errorf("插件ID不能为空")
	}
	if plugin.Name == "" {
		return nil, fmt.Errorf("插件名称不能为空")
	}
	if plugin.Type == "" {
		return nil, fmt.Errorf("插件类型不能为空")
	}

	// 验证插件类型
	validTypes := map[PluginType]bool{
		PluginTypeDatasource:   true,
		PluginTypeNotification: true,
		PluginTypeAI:           true,
	}
	if !validTypes[plugin.Type] {
		return nil, fmt.Errorf("不支持的插件类型: %s", plugin.Type)
	}

	// 检查ID是否已存在
	m.mu.RLock()
	_, exists := m.plugins[plugin.ID]
	m.mu.RUnlock()
	if exists {
		return nil, fmt.Errorf("插件ID已存在: %s", plugin.ID)
	}

	// 设置时间
	plugin.CreatedAt = time.Now()
	plugin.UpdatedAt = time.Now()

	// 添加插件
	if err := m.AddPlugin(&plugin); err != nil {
		return nil, err
	}

	return &plugin, nil
}

// ImportPluginFromFile 从文件导入插件
func (m *Manager) ImportPluginFromFile(filePath string) (*Plugin, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	return m.ImportPlugin(string(data))
}

// ExportPlugin 导出插件为JSON字符串
func (m *Manager) ExportPlugin(id string) (string, error) {
	plugin, err := m.GetPlugin(id)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(plugin, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化插件失败: %w", err)
	}

	return string(data), nil
}

// GetPluginsDir 获取插件目录路径
func (m *Manager) GetPluginsDir() string {
	return m.pluginsDir
}

// ScanPluginFiles 扫描插件目录中的JSON文件
func (m *Manager) ScanPluginFiles() ([]string, error) {
	var files []string

	entries, err := os.ReadDir(m.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("读取插件目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".json" && entry.Name() != "plugins.json" {
			files = append(files, filepath.Join(m.pluginsDir, entry.Name()))
		}
	}

	return files, nil
}

// ImportAllPluginFiles 导入插件目录中的所有JSON文件
func (m *Manager) ImportAllPluginFiles() (int, []error) {
	files, err := m.ScanPluginFiles()
	if err != nil {
		return 0, []error{err}
	}

	var imported int
	var errors []error

	for _, file := range files {
		_, err := m.ImportPluginFromFile(file)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", filepath.Base(file), err))
		} else {
			imported++
			// 导入成功后删除文件（已保存到plugins.json）
			os.Remove(file)
		}
	}

	return imported, errors
}
