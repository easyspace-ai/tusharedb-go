package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// FetchQuote 从数据源插件获取实时行情
func (m *Manager) FetchQuote(pluginID string, code string) (*DatasourceResult, error) {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	if plugin.Type != PluginTypeDatasource {
		return nil, fmt.Errorf("插件类型错误: %s", plugin.Type)
	}

	if !plugin.Enabled {
		return nil, fmt.Errorf("插件未启用: %s", plugin.Name)
	}

	var config DatasourceConfig
	if err := json.Unmarshal(plugin.Config, &config); err != nil {
		return nil, fmt.Errorf("解析数据源配置失败: %w", err)
	}

	// 构建请求URL
	endpoint := config.Endpoints.Quote
	if endpoint == "" {
		return nil, fmt.Errorf("未配置行情端点")
	}

	url := config.BaseURL + m.replaceCodeInURL(endpoint, code, config.Params)

	// 发送请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range config.Headers {
		// 替换参数
		for pk, pv := range config.Params {
			value = strings.ReplaceAll(value, "{"+pk+"}", pv)
		}
		req.Header.Set(key, value)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("请求失败: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 根据映射提取数据
	result := &DatasourceResult{}
	result.Price = m.extractFloat(respData, config.Mapping.Price)
	result.Change = m.extractFloat(respData, config.Mapping.Change)
	result.ChangePercent = m.extractFloat(respData, config.Mapping.ChangePercent)
	result.Volume = m.extractFloat(respData, config.Mapping.Volume)
	result.Amount = m.extractFloat(respData, config.Mapping.Amount)
	result.High = m.extractFloat(respData, config.Mapping.High)
	result.Low = m.extractFloat(respData, config.Mapping.Low)
	result.Open = m.extractFloat(respData, config.Mapping.Open)
	result.PreClose = m.extractFloat(respData, config.Mapping.PreClose)
	result.Name = m.extractString(respData, config.Mapping.Name)

	return result, nil
}

// FetchQuoteFromAll 从所有启用的数据源插件获取行情（返回第一个成功的）
func (m *Manager) FetchQuoteFromAll(code string) (*DatasourceResult, string, error) {
	plugins := m.GetPluginsByType(PluginTypeDatasource)

	for _, plugin := range plugins {
		if !plugin.Enabled {
			continue
		}

		result, err := m.FetchQuote(plugin.ID, code)
		if err == nil && result.Price > 0 {
			return result, plugin.ID, nil
		}
	}

	return nil, "", fmt.Errorf("所有数据源都无法获取数据")
}

// TestDatasource 测试数据源插件
func (m *Manager) TestDatasource(pluginID string, code string) (*DatasourceResult, error) {
	if code == "" {
		code = "000001" // 默认测试平安银行
	}
	return m.FetchQuote(pluginID, code)
}

// GetEnabledDatasourcePlugins 获取所有启用的数据源插件
func (m *Manager) GetEnabledDatasourcePlugins() []Plugin {
	plugins := m.GetPluginsByType(PluginTypeDatasource)
	var enabled []Plugin
	for _, p := range plugins {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled
}

// HasEnabledDatasourcePlugins 检查是否有启用的数据源插件
func (m *Manager) HasEnabledDatasourcePlugins() bool {
	return len(m.GetEnabledDatasourcePlugins()) > 0
}

// replaceCodeInURL 替换URL中的股票代码和参数
func (m *Manager) replaceCodeInURL(url string, code string, params map[string]string) string {
	result := strings.ReplaceAll(url, "{code}", code)
	for key, value := range params {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	return result
}

// extractFloat 从嵌套map中提取float值
func (m *Manager) extractFloat(data map[string]interface{}, path string) float64 {
	if path == "" {
		return 0
	}

	value := m.extractValue(data, path)
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

// extractString 从嵌套map中提取string值
func (m *Manager) extractString(data map[string]interface{}, path string) string {
	if path == "" {
		return ""
	}

	value := m.extractValue(data, path)
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// extractValue 从嵌套map中提取值（支持点号路径，如 "data.quote.price"）
func (m *Manager) extractValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		// 检查是否是数组索引，如 "items[0]"
		if idx := strings.Index(part, "["); idx != -1 {
			key := part[:idx]
			indexStr := part[idx+1 : len(part)-1]
			index, _ := strconv.Atoi(indexStr)

			if m, ok := current.(map[string]interface{}); ok {
				current = m[key]
			} else {
				return nil
			}

			if arr, ok := current.([]interface{}); ok {
				if index < len(arr) {
					current = arr[index]
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			if m, ok := current.(map[string]interface{}); ok {
				current = m[part]
			} else {
				return nil
			}
		}

		if current == nil {
			return nil
		}
	}

	return current
}

// 预置数据源模板
var DatasourceTemplates = []struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Config      DatasourceConfig `json:"config"`
}{
	{
		ID:          "custom-api",
		Name:        "自定义API",
		Description: "接入自定义的股票行情API",
		Config: DatasourceConfig{
			BaseURL: "",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Params: map[string]string{
				"apiKey": "",
			},
			Endpoints: DatasourceEndpoints{
				Quote: "/quote?code={code}&apiKey={apiKey}",
			},
			Mapping: DatasourceMapping{
				Price:        "data.price",
				Change:       "data.change",
				ChangePercent: "data.changePercent",
				Volume:       "data.volume",
				High:         "data.high",
				Low:          "data.low",
				Open:         "data.open",
				PreClose:     "data.preClose",
				Name:         "data.name",
			},
		},
	},
}
