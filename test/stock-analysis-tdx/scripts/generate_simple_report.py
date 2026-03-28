#!/usr/bin/env python3
"""
简单 HTML 报告生成脚本
"""

import argparse
import json
import os
import sys
from datetime import datetime
from jinja2 import Template


# 简化版 HTML 模板
HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ company_name }} - 股票分析报告</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', 'Microsoft YaHei', sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 15px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2E86AB 0%, #A23B72 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header .meta {
            display: flex; justify-content: center; gap: 30px;
            margin-top: 20px; flex-wrap: wrap;
        }
        .header .meta-item {
            background: rgba(255,255,255,0.2);
            padding: 10px 20px; border-radius: 25px;
        }
        .section { padding: 40px; border-bottom: 1px solid #eee; }
        .section:last-child { border-bottom: none; }
        .section-title {
            font-size: 1.8em; color: #2E86AB;
            margin-bottom: 25px; padding-bottom: 10px;
            border-bottom: 3px solid #2E86AB;
            display: flex; align-items: center;
        }
        .section-title::before {
            content: ''; display: inline-block;
            width: 8px; height: 30px; background: #2E86AB;
            margin-right: 15px; border-radius: 4px;
        }
        .core-conclusion {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white; padding: 30px; border-radius: 10px;
            margin-bottom: 30px;
        }
        .core-conclusion h2 { margin-bottom: 20px; font-size: 1.5em; }
        .conclusion-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px; margin-top: 20px;
        }
        .conclusion-item {
            background: rgba(255,255,255,0.2);
            padding: 20px; border-radius: 10px; text-align: center;
        }
        .conclusion-item .label { font-size: 0.9em; opacity: 0.9; margin-bottom: 10px; }
        .conclusion-item .value { font-size: 2em; font-weight: bold; }
        .charts-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
            gap: 30px;
        }
        .chart-container {
            background: #f8f9fa; padding: 20px;
            border-radius: 10px; text-align: center;
        }
        .chart-container h3 { color: #2E86AB; margin-bottom: 15px; }
        .chart-container img {
            max-width: 100%; height: auto;
            border-radius: 8px; box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .footer {
            background: #f8f9fa; padding: 30px;
            text-align: center; color: #666;
            border-top: 1px solid #e0e0e0;
        }
        .disclaimer {
            background: #fff3cd; border: 1px solid #ffc107;
            padding: 20px; border-radius: 8px;
            margin-top: 20px; text-align: left;
        }
        .disclaimer strong { color: #856404; }
        @media (max-width: 768px) {
            .charts-grid { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{ company_name }}</h1>
            <div class="meta">
                <div class="meta-item">股票代码: {{ symbol }}</div>
                <div class="meta-item">分析日期: {{ analysis_date }}</div>
                <div class="meta-item">当前价格: {{ current_price }}</div>
                <div class="meta-item">数据来源: TDX(通达信)</div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">核心结论</h2>
            <div class="core-conclusion">
                <h2>{{ one_line_summary }}</h2>
                <div class="conclusion-grid">
                    <div class="conclusion-item">
                        <div class="label">投资建议</div>
                        <div class="value">{{ recommendation }}</div>
                    </div>
                    <div class="conclusion-item">
                        <div class="label">预期收益</div>
                        <div class="value">{{ expected_return }}</div>
                    </div>
                    <div class="conclusion-item">
                        <div class="label">最大风险</div>
                        <div class="value">{{ max_risk }}</div>
                    </div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2 class="section-title">技术分析图表</h2>
            <div class="charts-grid">
                {% for chart in charts %}
                <div class="chart-container">
                    <h3>{{ chart_names[chart.type] if chart.type in chart_names else chart.type }}</h3>
                    <img src="charts/{{ chart.filename }}" alt="{{ chart.type }}">
                </div>
                {% endfor %}
            </div>
        </div>

        <div class="footer">
            <p>报告生成时间: {{ generation_time }}</p>
            <div class="disclaimer">
                <strong>免责声明：</strong>
                <p>本报告由AI分析系统基于TDX(通达信)数据生成，仅供参考，不构成任何投资建议。</p>
                <p>股票投资存在风险，市场有风险，投资需谨慎。</p>
            </div>
        </div>
    </div>
</body>
</html>
"""


def generate_html_report(data, output_path):
    try:
        template = Template(HTML_TEMPLATE)

        chart_names = {
            'price_chart': '股价走势图',
            'candlestick': 'K线图与成交量',
            'macd': 'MACD指标',
            'rsi': 'RSI指标'
        }

        template_data = {
            'company_name': data.get('company_name', 'N/A'),
            'symbol': data.get('symbol', 'N/A'),
            'analysis_date': data.get('analysis_date', datetime.now().strftime('%Y-%m-%d')),
            'current_price': data.get('current_price', 'N/A'),
            'one_line_summary': data.get('one_line_summary', '暂无总结'),
            'recommendation': data.get('recommendation', '持有'),
            'expected_return': data.get('expected_return', '+0%'),
            'max_risk': data.get('max_risk', '-0%'),
            'charts': data.get('charts', []),
            'chart_names': chart_names,
            'generation_time': datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        }

        html_content = template.render(**template_data)

        output_dir = os.path.dirname(output_path)
        if output_dir:
            os.makedirs(output_dir, exist_ok=True)

        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(html_content)

        return {
            'success': True,
            'file_path': output_path,
            'file_size': len(html_content)
        }

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {
            'success': False,
            'error': str(e)
        }


def main():
    parser = argparse.ArgumentParser(description='生成简单HTML报告')
    parser.add_argument('--data', type=str, required=True, help='报告数据JSON文件')
    parser.add_argument('--output', type=str, default='report.html', help='输出HTML文件路径')

    args = parser.parse_args()

    with open(args.data, 'r', encoding='utf-8') as f:
        data = json.load(f)

    result = generate_html_report(data, args.output)
    print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
