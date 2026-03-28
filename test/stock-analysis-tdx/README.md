# 股票分析团队 (TDX数据源版)

基于通达信(TDX)数据源的多Agent股票分析系统。

## 简介

本项目改造自"麻省理工学院"开源的多Agent股神框架，使用TDX（通达信）API替换原有数据源，提供更稳定、更快速的A股数据获取能力。

系统模拟人类金融团队的运作模式，包含：
- **分析师团队**：基本面、情绪、新闻、技术分析师
- **研究员团队**：看多、看空研究员（辩论机制）
- **交易员团队**：综合研判，制定交易计划
- **风控与执行团队**：风险评估，最终决策

## 项目结构

```
stock-analysis-tdx/
├── SKILL.md                      # Skill配置文档
├── README.md                     # 本文件
├── scripts/
│   ├── tdx_data_fetcher.py       # TDX数据获取脚本（新增）
│   ├── chart_generator.py        # 图表生成脚本
│   ├── html_report_generator.py  # HTML报告生成
│   ├── enhanced_html_report_generator.py  # 增强版报告
│   └── image_fetcher.py          # 图片获取
└── references/
    ├── analysis-framework.md      # 分析框架
    ├── market-strategy.md         # 市场策略
    ├── risk-scoring-criteria.md   # 风险评分标准
    ├── backtesting-guidelines.md  # 回测指南
    ├── chart-guide.md             # 图表指南
    ├── report-template.md         # 报告模板
    └── html-template.html         # HTML模板
```

## 快速开始

### 前置要求

1. **启动TDX API服务**

```bash
cd /Users/leven/space/hein/stockdb/test/tdx-api/web
go run .
```

服务默认运行在 `http://localhost:8080`

2. **安装Python依赖**

```bash
pip install pandas numpy ta matplotlib requests jinja2 plotly
```

### 使用示例

#### 1. 获取单只股票数据

```bash
python scripts/tdx_data_fetcher.py --symbol 000001 --market cn
```

#### 2. 获取市场指数数据

```bash
python scripts/tdx_data_fetcher.py --mode market --market cn
```

#### 3. 同时获取个股和大盘数据

```bash
python3 scripts/tdx_data_fetcher.py --symbol 000001 --market cn --mode both
```

## 数据来源

### TDX API 接口

本系统使用以下TDX API接口：

| 接口 | 说明 |
|-----|------|
| `/quote` | 五档行情（实时价格、成交量） |
| `/kline` | K线数据（支持多种周期） |
| `/kline-all/ths` | 同花顺前复权K线 |
| `/minute` | 分时数据 |
| `/search` | 股票搜索 |
| `/stock-info` | 综合信息 |

### 数据单位说明

- **价格**：TDX返回单位为"厘"，脚本自动转换为"元"（1元 = 1000厘）
- **成交量**：TDX返回单位为"手"，脚本自动转换为"股"（1手 = 100股）

## 核心Agent团队

### 1. 分析师团队

| 角色 | 职责 |
|-----|------|
| 基本面分析师 | 评估财务状况和业绩指标 |
| 情绪分析师 | 分析市场情绪和舆情 |
| 新闻分析师 | 解读新闻和宏观经济影响 |
| 技术分析师 | 利用技术指标预测走势 |

### 2. 研究员团队

| 角色 | 职责 |
|-----|------|
| 看多研究员 | 竭尽全力寻找买入理由 |
| 看空研究员 | 竭尽全力寻找卖出理由 |

通过**辩论机制**消除偏见，平衡收益与风险。

### 3. 交易员团队

综合分析师和研究员报告，做出明智的交易决策。

### 4. 风控与执行团队

评估风险等级（1-10分），投资组合经理拥有最终决定权。

## 风险等级说明

| 分数 | 等级 | 说明 |
|-----|------|------|
| 1-3 | 低风险 | 适合稳健投资者 |
| 4-6 | 中等风险 | 适合平衡型投资者 |
| 7-10 | 高风险 | 适合激进型投资者 |

## 故障排除

### TDX API 连接失败

**问题**：`无法连接到 TDX API 服务`

**解决方案**：

1. 检查服务是否启动：
```bash
curl http://localhost:8080/health
```

2. 如果服务未启动，进入tdx-api目录启动：
```bash
cd /Users/leven/space/hein/stockdb/test/tdx-api/web
go run .
```

3. 检查防火墙设置

4. 使用环境变量指定其他地址：
```bash
export TDX_API_BASE_URL=http://other-host:8080
```

### 数据获取失败

**问题**：`无法获取股票数据`

**解决方案**：

1. 确认股票代码格式正确（6位数字，如 000001）
2. 指数代码需要带前缀：
   - 上证指数：`sh000001`
   - 深证成指：`sz399001`
3. 检查TDX API服务日志

## 与原版本的区别

| 特性 | 原版本 | TDX版本 |
|-----|--------|---------|
| 数据源 | Yahoo Finance + Akshare | TDX (通达信) |
| 数据稳定性 | 依赖境外服务，不稳定 | 境内数据源，稳定 |
| 前复权支持 | 部分支持 | 完整支持（同花顺源） |
| 实时行情 | 延迟较高 | 实时五档 |
| A股支持 | 需要转换 | 原生支持 |

## 免责声明

本系统仅供学习和研究使用，不构成任何投资建议。股市有风险，投资需谨慎。

## 致谢

- 原项目：stock-analysis-team
- TDX API：基于 [injoyai/tdx](https://github.com/injoyai/tdx)
