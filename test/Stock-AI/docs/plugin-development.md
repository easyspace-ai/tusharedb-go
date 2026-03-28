# 插件系统开发文档

## 概述

Stock AI 提供开放的插件系统，允许第三方开发者编写自定义插件来扩展软件功能。

## 插件类型

| 类型 | 说明 | 状态 |
|------|------|------|
| notification | 通知插件 - 推送消息到各种渠道 | ✅ 可用 |
| datasource | 数据源插件 - 接入自定义行情API | ✅ 可用 |
| ai | AI模型插件 - 接入自定义AI模型 | ✅ 可用 |
| indicator | 指标插件 - AI提示词分析技术指标 | ✅ 可用 |

## 快速开始

### 插件目录

插件存放在用户目录下：
- Windows: `C:\Users\<用户名>\.stock-ai\plugins\`
- macOS: `~/.stock-ai/plugins/`
- Linux: `~/.stock-ai/plugins/`

### 插件结构

每个插件是一个 JSON 文件，基本结构如下：

```json
{
  "id": "my-plugin-001",
  "name": "我的插件",
  "type": "notification",
  "version": "1.0.0",
  "author": "开发者名称",
  "description": "插件描述",
  "homepage": "https://github.com/xxx/xxx",
  "enabled": true,
  "config": {
    // 插件配置，根据类型不同而不同
  }
}
```

## 通知插件开发

通知插件用于在股票提醒触发时，将消息推送到指定渠道。

### 配置结构

```json
{
  "id": "my-notification",
  "name": "我的通知插件",
  "type": "notification",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "自定义通知渠道",
  "enabled": true,
  "config": {
    "url": "https://api.example.com/send",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-token"
    },
    "contentType": "json",
    "bodyTemplate": {
      "title": "股票提醒: {stockName}",
      "content": "{message}",
      "extra": {
        "code": "{stockCode}",
        "price": "{currentPrice}"
      }
    }
  }
}
```

### 配置字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| url | string | ✅ | 请求地址，支持变量替换 |
| method | string | ❌ | HTTP方法，默认 POST |
| headers | object | ❌ | 请求头 |
| contentType | string | ❌ | 内容类型：json/form/text |
| bodyTemplate | object/string | ❌ | 请求体模板 |
| params | object | ❌ | 自定义参数 |

### 可用变量

在 `url`、`headers`、`bodyTemplate` 中可以使用以下变量：

| 变量 | 说明 | 示例值 |
|------|------|--------|
| {stockCode} | 股票代码 | 000001 |
| {stockName} | 股票名称 | 平安银行 |
| {alertType} | 提醒类型 | 股价提醒/涨跌提醒 |
| {currentPrice} | 当前价格 | 10.50 |
| {condition} | 触发条件 | 高于/低于 |
| {targetValue} | 目标值 | 10.00 |
| {triggerTime} | 触发时间 | 2024-01-15 10:30:00 |
| {change} | 涨跌额 | 0.25 |
| {changePercent} | 涨跌幅 | 2.44% |
| {message} | 完整消息 | 平安银行 股价已达到 10.50 元 |

### 示例：Telegram 机器人

```json
{
  "id": "telegram-bot",
  "name": "Telegram 机器人",
  "type": "notification",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "通过 Telegram Bot 发送通知",
  "enabled": true,
  "config": {
    "url": "https://api.telegram.org/bot{botToken}/sendMessage",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "params": {
      "botToken": "你的Bot Token",
      "chatId": "你的Chat ID"
    },
    "contentType": "json",
    "bodyTemplate": {
      "chat_id": "{chatId}",
      "text": "📈 *{stockName}* ({stockCode})\n\n{alertType}\n当前价格: {currentPrice}\n触发条件: {condition} {targetValue}\n时间: {triggerTime}",
      "parse_mode": "Markdown"
    }
  }
}
```

### 示例：邮件通知（通过 SMTP API）

```json
{
  "id": "email-notify",
  "name": "邮件通知",
  "type": "notification",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "通过邮件发送通知",
  "enabled": true,
  "config": {
    "url": "https://api.sendgrid.com/v3/mail/send",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json",
      "Authorization": "Bearer {apiKey}"
    },
    "params": {
      "apiKey": "你的SendGrid API Key",
      "toEmail": "收件人邮箱",
      "fromEmail": "发件人邮箱"
    },
    "contentType": "json",
    "bodyTemplate": {
      "personalizations": [{"to": [{"email": "{toEmail}"}]}],
      "from": {"email": "{fromEmail}"},
      "subject": "股票提醒: {stockName}",
      "content": [{"type": "text/plain", "value": "{message}"}]
    }
  }
}
```

### 示例：Discord Webhook

```json
{
  "id": "discord-webhook",
  "name": "Discord 通知",
  "type": "notification",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "通过 Discord Webhook 发送通知",
  "enabled": true,
  "config": {
    "url": "https://discord.com/webhooks/xxx/xxx",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "contentType": "json",
    "bodyTemplate": {
      "content": "📈 **{stockName}** ({stockCode})",
      "embeds": [{
        "title": "{alertType}",
        "description": "当前价格: {currentPrice}\n触发条件: {condition} {targetValue}",
        "color": 5814783,
        "footer": {"text": "{triggerTime}"}
      }]
    }
  }
}
```

### 示例：Slack Webhook

```json
{
  "id": "slack-webhook",
  "name": "Slack 通知",
  "type": "notification",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "通过 Slack Webhook 发送通知",
  "enabled": true,
  "config": {
    "url": "https://hooks.slack.com/services/xxx/xxx/xxx",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "contentType": "json",
    "bodyTemplate": {
      "text": "股票提醒",
      "blocks": [
        {
          "type": "header",
          "text": {"type": "plain_text", "text": "📈 {stockName} ({stockCode})"}
        },
        {
          "type": "section",
          "fields": [
            {"type": "mrkdwn", "text": "*类型:*\n{alertType}"},
            {"type": "mrkdwn", "text": "*当前价格:*\n{currentPrice}"},
            {"type": "mrkdwn", "text": "*触发条件:*\n{condition} {targetValue}"},
            {"type": "mrkdwn", "text": "*时间:*\n{triggerTime}"}
          ]
        }
      ]
    }
  }
}
```

## 数据源插件开发

数据源插件用于接入自定义的股票行情API。

### 配置结构

```json
{
  "id": "my-datasource",
  "name": "我的数据源",
  "type": "datasource",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "自定义行情数据源",
  "enabled": true,
  "config": {
    "baseUrl": "https://api.example.com",
    "endpoints": {
      "quote": "/stock/quote?code={code}"
    },
    "headers": {
      "Authorization": "Bearer {apiKey}"
    },
    "params": {
      "apiKey": "你的API Key"
    },
    "mapping": {
      "price": "data.current",
      "change": "data.change",
      "changePercent": "data.percent",
      "volume": "data.volume",
      "high": "data.high",
      "low": "data.low",
      "open": "data.open",
      "preClose": "data.last_close",
      "name": "data.name"
    }
  }
}
```

### 字段映射说明

`mapping` 用于将API返回的数据映射到标准字段。使用点号表示嵌套路径，支持数组索引：

```json
{
  "price": "data.current",      // 从 response.data.current 获取
  "change": "result[0].change"  // 从 response.result[0].change 获取
}
```

### 标准字段

| 字段 | 说明 |
|------|------|
| price | 当前价格 |
| change | 涨跌额 |
| changePercent | 涨跌幅 |
| volume | 成交量 |
| amount | 成交额 |
| high | 最高价 |
| low | 最低价 |
| open | 开盘价 |
| preClose | 昨收价 |
| name | 股票名称 |

## AI模型插件开发

AI模型插件用于接入自定义的AI大模型，支持OpenAI兼容接口。

### 配置结构

```json
{
  "id": "my-ai-model",
  "name": "我的AI模型",
  "type": "ai",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "自定义AI模型",
  "enabled": true,
  "config": {
    "provider": "openai-compatible",
    "baseUrl": "https://api.example.com/v1",
    "apiKey": "你的API Key",
    "model": "gpt-4",
    "maxTokens": 4096,
    "temperature": 0.7,
    "systemPrompt": "你是一个专业的股票分析师...",
    "headers": {}
  }
}
```

### 配置字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| provider | string | 是 | 固定为 "openai-compatible" |
| baseUrl | string | 是 | API基础地址 |
| apiKey | string | 是 | API密钥 |
| model | string | 是 | 模型名称 |
| maxTokens | int | 否 | 最大token数，默认4096 |
| temperature | float | 否 | 温度参数，默认0.7 |
| systemPrompt | string | 否 | 系统提示词 |
| headers | object | 否 | 额外的请求头 |

### 支持的AI服务

| 服务 | baseUrl | 说明 |
|------|---------|------|
| DeepSeek | https://api.deepseek.com/v1 | 性价比高，推荐 |
| 通义千问 | https://dashscope.aliyuncs.com/compatible-mode/v1 | 阿里云 |
| 智谱GLM | https://open.bigmodel.cn/paas/v4 | 清华 |
| 硅基流动 | https://api.siliconflow.cn/v1 | 聚合多模型 |
| Ollama | http://localhost:11434/v1 | 本地部署 |

### 示例：DeepSeek

```json
{
  "id": "deepseek-chat",
  "name": "DeepSeek Chat",
  "type": "ai",
  "version": "1.0.0",
  "description": "DeepSeek AI模型",
  "enabled": true,
  "config": {
    "provider": "openai-compatible",
    "baseUrl": "https://api.deepseek.com/v1",
    "apiKey": "sk-xxx",
    "model": "deepseek-chat",
    "maxTokens": 4096,
    "temperature": 0.7,
    "systemPrompt": "你是一个专业的股票分析师，擅长技术分析和基本面分析。"
  }
}
```

### 示例：Ollama本地模型

```json
{
  "id": "ollama-qwen",
  "name": "Ollama Qwen",
  "type": "ai",
  "version": "1.0.0",
  "description": "本地Ollama模型",
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

## 插件安装

### 方法一：手动安装

1. 下载插件 JSON 文件
2. 将文件复制到插件目录
3. 重启软件或在插件管理页面刷新

### 方法二：通过软件安装

1. 打开软件 → 插件管理
2. 点击"导入插件"
3. 选择插件 JSON 文件

## 插件开发规范

### 命名规范

- **id**: 使用小写字母、数字和连字符，如 `my-plugin-001`
- **name**: 简洁明了的中文或英文名称
- **version**: 遵循语义化版本，如 `1.0.0`

### 安全规范

1. **不要在插件中硬编码敏感信息**，使用 `params` 字段让用户配置
2. **URL 必须使用 HTTPS**（本地开发除外）
3. **不要请求不必要的权限**

### 最佳实践

1. 提供清晰的 `description` 说明插件功能
2. 在 `homepage` 提供文档链接
3. 使用有意义的变量名
4. 测试各种边界情况

## 插件分享

欢迎将你开发的插件分享到社区：

1. 在 GitHub 创建仓库存放插件
2. 提交 PR 到官方插件仓库
3. 在 Issues 中分享你的插件

## 常见问题

### Q: 插件不生效怎么办？

1. 检查 JSON 格式是否正确
2. 检查 `enabled` 是否为 `true`
3. 检查 URL 是否可访问
4. 查看软件日志获取错误信息

### Q: 如何调试插件？

1. 使用"测试"按钮发送测试消息
2. 检查目标平台是否收到消息
3. 如果失败，检查请求格式是否正确

### Q: 变量没有被替换？

1. 确保变量名拼写正确（区分大小写）
2. 确保使用花括号包裹：`{variableName}`
3. 检查变量是否在支持的列表中

## 更新日志

### v1.1.0 (2025-01)
- 新增指标插件（基于AI提示词）
- 数据源插件正式可用
- AI模型插件正式可用
- 策略插件功能已移除（迁移至独立策略/回测产品）

### v1.0.0 (2024-01)
- 初始版本
- 支持通知插件
- 预置钉钉、企业微信、飞书等模板

## 指标插件开发

指标插件使用AI提示词来分析股票技术指标，无需编写复杂的脚本代码。

### 配置结构

```json
{
  "id": "my-indicator",
  "name": "我的指标插件",
  "type": "indicator",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "基于AI的技术指标分析",
  "enabled": true,
  "config": {
    "prompt": "请分析以下股票的MACD指标...",
    "aiPluginId": "",
    "outputFormat": "signal",
    "params": {}
  }
}
```

### 配置字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| prompt | string | ✅ | AI提示词模板，支持变量替换 |
| aiPluginId | string | ❌ | 指定使用的AI插件ID，为空则使用默认 |
| outputFormat | string | ❌ | 输出格式：signal/value/text |
| params | object | ❌ | 自定义参数 |

### 可用变量

在 `prompt` 中可以使用以下变量：

| 变量 | 说明 | 示例值 |
|------|------|--------|
| {code} | 股票代码 | sz000001 |
| {name} | 股票名称 | 平安银行 |
| {price} | 当前价格 | 10.50 |
| {change} | 涨跌额 | 0.25 |
| {changePercent} | 涨跌幅 | 2.44% |
| {volume} | 成交量 | 100000000 |
| {amount} | 成交额 | 1050000000 |
| {high} | 最高价 | 10.80 |
| {low} | 最低价 | 10.20 |
| {open} | 开盘价 | 10.30 |
| {preClose} | 昨收价 | 10.25 |
| {klines} | K线数据（CSV格式） | 日期,开盘,收盘... |

### 示例：MACD分析插件

```json
{
  "id": "macd-analysis",
  "name": "MACD趋势分析",
  "type": "indicator",
  "version": "1.0.0",
  "description": "基于AI分析MACD指标，判断买卖信号",
  "enabled": true,
  "config": {
    "prompt": "请分析以下股票的MACD指标趋势：\n\n股票：{name}（{code}）\n当前价格：{price}\n涨跌幅：{changePercent}\n\n近期K线数据：\n{klines}\n\n请根据K线数据计算MACD指标，并给出：\n1. 当前MACD值和信号线位置\n2. 是否出现金叉或死叉\n3. 买卖建议（买入/卖出/观望）\n\n请在回答开头明确标注：【买入】【卖出】或【观望】",
    "outputFormat": "signal"
  }
}
```

### 示例：KDJ超买超卖分析

```json
{
  "id": "kdj-analysis",
  "name": "KDJ超买超卖分析",
  "type": "indicator",
  "version": "1.0.0",
  "description": "基于AI分析KDJ指标，判断超买超卖",
  "enabled": true,
  "config": {
    "prompt": "请分析以下股票的KDJ指标：\n\n股票：{name}（{code}）\n当前价格：{price}\n今日最高：{high}\n今日最低：{low}\n\n近期K线数据：\n{klines}\n\n请分析：\n1. 当前K、D、J值\n2. 是否处于超买（>80）或超卖（<20）区域\n3. 买卖建议\n\n请在回答开头明确标注：【买入】【卖出】或【观望】",
    "outputFormat": "signal"
  }
}
```
