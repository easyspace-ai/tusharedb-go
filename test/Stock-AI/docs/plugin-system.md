# 插件系统设计文档

## 概述

本文档描述 stock-ai 的插件系统设计方案，旨在让用户能够扩展软件功能，同时为定制化服务提供基础。

## 插件类型

### 1. 数据源插件

允许用户接入自定义的行情数据源。

**配置格式：**
```json
{
  "name": "自定义数据源",
  "type": "datasource",
  "enabled": true,
  "config": {
    "url": "https://api.example.com/stock/{code}",
    "method": "GET",
    "headers": {
      "Authorization": "Bearer {token}"
    },
    "params": {
      "token": "用户的API Token"
    },
    "responseMapping": {
      "price": "data.current",
      "change": "data.change",
      "changePercent": "data.change_percent",
      "volume": "data.volume",
      "high": "data.high",
      "low": "data.low",
      "open": "data.open",
      "preClose": "data.pre_close"
    }
  }
}
```

**支持的功能：**
- 实时行情获取
- K线数据获取
- 分时数据获取

### 2. 通知插件

允许用户配置自定义通知渠道。

**配置格式：**
```json
{
  "name": "钉钉通知",
  "type": "notification",
  "enabled": true,
  "config": {
    "url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "bodyTemplate": {
      "msgtype": "text",
      "text": {
        "content": "【股票提醒】{stockName}({stockCode}) {alertType}: 当前价格 {currentPrice}，触发条件 {condition} {targetValue}"
      }
    }
  }
}
```

**支持的通知渠道：**
- Webhook（通用）
- 钉钉机器人
- 企业微信机器人
- 飞书机器人
- Server酱
- Bark（iOS推送）
- 自定义HTTP接口

### 3. AI模型插件

允许用户接入自定义的AI模型。

**配置格式：**
```json
{
  "name": "本地Ollama",
  "type": "ai",
  "enabled": true,
  "config": {
    "provider": "openai-compatible",
    "baseUrl": "http://localhost:11434/v1",
    "apiKey": "",
    "model": "qwen2.5:7b",
    "maxTokens": 4096,
    "temperature": 0.7
  }
}
```

**支持的AI提供商：**
- OpenAI兼容接口（Ollama、LocalAI、vLLM等）
- 自定义HTTP接口

### 4. 指标插件

允许用户编写自定义技术指标。

**配置格式：**
```json
{
  "name": "自定义MACD",
  "type": "indicator",
  "enabled": true,
  "config": {
    "script": "indicator.js",
    "params": {
      "fast": 12,
      "slow": 26,
      "signal": 9
    }
  }
}
```

**指标脚本示例 (JavaScript)：**
```javascript
// indicator.js
function calculate(klines, params) {
  const { fast, slow, signal } = params;
  // 计算EMA
  function ema(data, period) {
    const k = 2 / (period + 1);
    let emaVal = data[0];
    const result = [emaVal];
    for (let i = 1; i < data.length; i++) {
      emaVal = data[i] * k + emaVal * (1 - k);
      result.push(emaVal);
    }
    return result;
  }

  const closes = klines.map(k => k.close);
  const emaFast = ema(closes, fast);
  const emaSlow = ema(closes, slow);
  const dif = emaFast.map((v, i) => v - emaSlow[i]);
  const dea = ema(dif, signal);
  const macd = dif.map((v, i) => (v - dea[i]) * 2);

  return { dif, dea, macd };
}
```

## 插件管理

### 插件目录结构

```
stock-ai/
├── plugins/
│   ├── config.json          # 插件配置文件
│   ├── datasources/         # 数据源插件
│   │   └── example.json
│   ├── notifications/       # 通知插件
│   │   ├── dingtalk.json
│   │   ├── wechat.json
│   │   └── bark.json
│   ├── ai/                  # AI模型插件
│   │   └── ollama.json
│   ├── indicators/          # 指标插件
│   │   ├── config.json
│   │   └── scripts/
│   │       └── macd.js
```

### 插件配置文件 (plugins/config.json)

```json
{
  "version": "1.0",
  "plugins": [
    {
      "id": "datasource-custom",
      "type": "datasource",
      "path": "datasources/example.json",
      "enabled": true
    },
    {
      "id": "notify-dingtalk",
      "type": "notification",
      "path": "notifications/dingtalk.json",
      "enabled": true
    },
    {
      "id": "ai-ollama",
      "type": "ai",
      "path": "ai/ollama.json",
      "enabled": false
    },
    {
      "id": "indicator-custom-macd",
      "type": "indicator",
      "path": "indicators/config.json",
      "enabled": true
    }
  ]
}
```

## 实现计划

### 第一阶段：基础框架

1. **插件管理器**
   - 插件加载和初始化
   - 插件配置读取和保存
   - 插件启用/禁用

2. **插件接口定义**
   - 定义各类插件的Go接口
   - 实现插件生命周期管理

3. **前端插件管理页面**
   - 插件列表展示
   - 插件配置编辑
   - 插件启用/禁用开关

### 第二阶段：通知插件

1. **Webhook通知**
   - 通用HTTP请求发送
   - 模板变量替换
   - 请求结果处理

2. **预置通知模板**
   - 钉钉机器人
   - 企业微信机器人
   - 飞书机器人
   - Bark推送

3. **通知测试功能**
   - 发送测试消息
   - 查看发送结果

### 第三阶段：数据源插件

1. **自定义数据源**
   - HTTP请求配置
   - 响应数据映射
   - 错误处理

2. **数据源优先级**
   - 多数据源切换
   - 失败自动降级

### 第四阶段：AI模型插件

1. **OpenAI兼容接口**
   - 支持Ollama、LocalAI等
   - 自定义模型参数

2. **模型切换**
   - 在设置中选择AI模型
   - 支持多模型配置

### 第五阶段：指标插件

1. **JavaScript引擎集成**
   - 使用goja或otto引擎
   - 安全沙箱执行

2. **指标计算**
   - 自定义指标脚本
   - 指标结果展示

## 安全考虑

1. **脚本沙箱**
   - JavaScript脚本在沙箱中执行
   - 限制文件系统和网络访问
   - 执行超时控制

2. **配置验证**
   - URL白名单（可选）
   - 敏感信息加密存储

3. **权限控制**
   - 插件权限声明
   - 用户授权确认

## API设计

### 后端API

```go
// 插件管理
func (a *App) GetPlugins() ([]Plugin, error)
func (a *App) GetPlugin(id string) (*Plugin, error)
func (a *App) SavePlugin(plugin *Plugin) error
func (a *App) DeletePlugin(id string) error
func (a *App) TogglePlugin(id string, enabled bool) error

// 通知插件
func (a *App) TestNotification(pluginId string) error
func (a *App) SendNotification(pluginId string, data map[string]interface{}) error

// 数据源插件
func (a *App) TestDatasource(pluginId string, code string) (*StockPrice, error)

// 指标插件
func (a *App) CalculateIndicator(pluginId string, code string) (map[string]interface{}, error)
```

### 前端API

```typescript
// 插件管理
export function GetPlugins(): Promise<Plugin[]>
export function GetPlugin(id: string): Promise<Plugin>
export function SavePlugin(plugin: Plugin): Promise<void>
export function DeletePlugin(id: string): Promise<void>
export function TogglePlugin(id: string, enabled: boolean): Promise<void>

// 通知插件
export function TestNotification(pluginId: string): Promise<void>

// 数据源插件
export function TestDatasource(pluginId: string, code: string): Promise<StockPrice>

// 指标插件
export function CalculateIndicator(pluginId: string, code: string): Promise<any>
```

## 预置插件模板

### 通知插件模板

#### 钉钉机器人
```json
{
  "name": "钉钉机器人",
  "type": "notification",
  "enabled": false,
  "config": {
    "url": "",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "bodyTemplate": {
      "msgtype": "markdown",
      "markdown": {
        "title": "股票提醒",
        "text": "### {stockName}({stockCode})\n\n**{alertType}**\n\n- 当前价格: {currentPrice}\n- 触发条件: {condition} {targetValue}\n- 触发时间: {triggerTime}"
      }
    }
  }
}
```

#### 企业微信机器人
```json
{
  "name": "企业微信机器人",
  "type": "notification",
  "enabled": false,
  "config": {
    "url": "",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "bodyTemplate": {
      "msgtype": "markdown",
      "markdown": {
        "content": "### 股票提醒\n**{stockName}**({stockCode})\n> {alertType}\n> 当前价格: <font color=\"warning\">{currentPrice}</font>\n> 触发条件: {condition} {targetValue}"
      }
    }
  }
}
```

#### Bark推送
```json
{
  "name": "Bark推送",
  "type": "notification",
  "enabled": false,
  "config": {
    "url": "https://api.day.app/{deviceKey}/{title}/{body}",
    "method": "GET",
    "params": {
      "deviceKey": "",
      "title": "股票提醒: {stockName}",
      "body": "{alertType} - 当前价格: {currentPrice}"
    }
  }
}
```

#### Server酱
```json
{
  "name": "Server酱",
  "type": "notification",
  "enabled": false,
  "config": {
    "url": "https://sctapi.ftqq.com/{sendKey}.send",
    "method": "POST",
    "headers": {
      "Content-Type": "application/x-www-form-urlencoded"
    },
    "params": {
      "sendKey": ""
    },
    "bodyTemplate": "title=股票提醒: {stockName}&desp={alertType}，当前价格: {currentPrice}，触发条件: {condition} {targetValue}"
  }
}
```

## 用户界面设计

### 插件管理页面

1. **插件列表**
   - 显示所有已安装插件
   - 插件类型图标
   - 启用/禁用开关
   - 编辑/删除按钮

2. **添加插件**
   - 选择插件类型
   - 选择预置模板或自定义
   - 填写配置信息

3. **插件配置**
   - 根据插件类型显示不同配置项
   - 配置验证
   - 测试功能

### 设置页面集成

在现有设置页面添加"插件"选项卡，或创建独立的插件管理页面。
