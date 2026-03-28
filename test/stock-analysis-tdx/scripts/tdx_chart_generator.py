#!/usr/bin/env python3
"""
TDX 股票图表生成脚本（基于通达信数据）

功能：
- 生成股价走势折线图
- 生成K线图（蜡烛图）
- 生成成交量柱状图
- 生成技术指标图（MACD、RSI）
- 使用 TDX 数据而非 yfinance
"""

import argparse
import json
import sys
import os
from datetime import datetime
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import numpy as np

# 禁用代理（和 tdx_data_fetcher.py 保持一致）
os.environ['NO_PROXY'] = '*'
os.environ['no_proxy'] = '*'
os.environ.pop('HTTP_PROXY', None)
os.environ.pop('http_proxy', None)
os.environ.pop('HTTPS_PROXY', None)
os.environ.pop('https_proxy', None)
os.environ.pop('ALL_PROXY', None)
os.environ.pop('all_proxy', None)

# 设置中文字体支持
plt.rcParams['font.sans-serif'] = ['DejaVu Sans', 'Arial Unicode MS', 'SimHei']
plt.rcParams['axes.unicode_minus'] = False

# 导入 tdx_data_fetcher
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, SCRIPT_DIR)
import tdx_data_fetcher


def generate_price_chart_from_df(df, symbol, company_name, output_dir="."):
    """
    从 DataFrame 生成股价走势折线图
    """
    try:
        # 创建图表
        fig, ax = plt.subplots(figsize=(12, 6))

        # 绘制股价线
        ax.plot(df.index, df['Close'], label='Close Price', linewidth=2, color='#2E86AB')

        # 绘制移动平均线
        if 'MA5' in df.columns:
            ax.plot(df.index, df['MA5'], label='MA5', linewidth=1, color='#F18F01', alpha=0.8)
        if 'MA20' in df.columns:
            ax.plot(df.index, df['MA20'], label='MA20', linewidth=1, color='#C73E1D', alpha=0.8)
        if 'MA60' in df.columns:
            ax.plot(df.index, df['MA60'], label='MA60', linewidth=1, color='#3B1F2B', alpha=0.8)

        # 设置标题和标签
        ax.set_title(f'{company_name} ({symbol}) Stock Price Trend', fontsize=16, fontweight='bold')
        ax.set_xlabel('Date', fontsize=12)
        ax.set_ylabel('Price (CNY)', fontsize=12)
        ax.legend(loc='best')
        ax.grid(True, alpha=0.3)

        # 格式化x轴
        ax.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))
        plt.xticks(rotation=45)

        plt.tight_layout()

        # 保存图表
        output_path = os.path.join(output_dir, f'{symbol}_price_chart.png')
        plt.savefig(output_path, dpi=150, bbox_inches='tight')
        plt.close()

        return {
            "success": True,
            "chart_type": "price_trend",
            "file_path": output_path,
            "filename": f'{symbol}_price_chart.png',
            "symbol": symbol
        }

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": f"生成价格图表失败: {str(e)}"}


def generate_candlestick_chart_from_df(df, symbol, company_name, output_dir="."):
    """
    从 DataFrame 生成K线图（蜡烛图）
    """
    try:
        # 取最近3个月数据
        if len(df) > 60:
            df_slice = df.tail(60).copy()
        else:
            df_slice = df.copy()

        # 创建图表
        fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10), gridspec_kw={'height_ratios': [3, 1]})

        # 绘制K线
        width = 0.6
        width2 = 0.1

        up = df_slice[df_slice['Close'] >= df_slice['Open']]
        down = df_slice[df_slice['Close'] < df_slice['Open']]

        # 绘制上涨的K线
        ax1.bar(up.index, up['Close'] - up['Open'], width, bottom=up['Open'], color='#00C853', alpha=0.7)
        ax1.bar(up.index, up['High'] - up['Close'], width2, bottom=up['Close'], color='#00C853', alpha=0.7)
        ax1.bar(up.index, up['Low'] - up['Open'], width2, bottom=up['Open'], color='#00C853', alpha=0.7)

        # 绘制下跌的K线
        ax1.bar(down.index, down['Close'] - down['Open'], width, bottom=down['Open'], color='#D50000', alpha=0.7)
        ax1.bar(down.index, down['High'] - down['Open'], width2, bottom=down['Open'], color='#D50000', alpha=0.7)
        ax1.bar(down.index, down['Low'] - down['Close'], width2, bottom=down['Close'], color='#D50000', alpha=0.7)

        ax1.set_title(f'{company_name} ({symbol}) Candlestick Chart', fontsize=16, fontweight='bold')
        ax1.set_ylabel('Price (CNY)', fontsize=12)
        ax1.grid(True, alpha=0.3)
        ax1.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))

        # 绘制成交量
        colors = ['#00C853' if close >= open_ else '#D50000'
                 for close, open_ in zip(df_slice['Close'], df_slice['Open'])]
        ax2.bar(df_slice.index, df_slice['Volume'], width=0.6, color=colors, alpha=0.7)
        ax2.set_ylabel('Volume', fontsize=12)
        ax2.set_xlabel('Date', fontsize=12)
        ax2.grid(True, alpha=0.3)
        ax2.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))

        plt.tight_layout()

        # 保存图表
        output_path = os.path.join(output_dir, f'{symbol}_candlestick.png')
        plt.savefig(output_path, dpi=150, bbox_inches='tight')
        plt.close()

        return {
            "success": True,
            "chart_type": "candlestick",
            "file_path": output_path,
            "filename": f'{symbol}_candlestick.png',
            "symbol": symbol
        }

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": f"生成K线图失败: {str(e)}"}


def generate_macd_chart_from_df(df, symbol, company_name, output_dir="."):
    """
    从 DataFrame 生成MACD指标图
    """
    try:
        # 创建图表
        fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10), gridspec_kw={'height_ratios': [2, 1]})

        # 绘制价格
        ax1.plot(df.index, df['Close'], label='Close Price', linewidth=2, color='#2E86AB')
        ax1.set_title(f'{company_name} ({symbol}) Price & MACD', fontsize=16, fontweight='bold')
        ax1.set_ylabel('Price (CNY)', fontsize=12)
        ax1.legend(loc='best')
        ax1.grid(True, alpha=0.3)
        ax1.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))

        # 绘制MACD
        if 'MACD' in df.columns:
            ax2.plot(df.index, df['MACD'], label='MACD', linewidth=1.5, color='#2196F3')
        if 'MACD_Signal' in df.columns:
            ax2.plot(df.index, df['MACD_Signal'], label='Signal', linewidth=1.5, color='#FF9800')
        if 'MACD_Diff' in df.columns:
            # 绘制柱状图
            colors = ['#00C853' if h >= 0 else '#D50000' for h in df['MACD_Diff']]
            ax2.bar(df.index, df['MACD_Diff'], width=0.8, color=colors, alpha=0.5)

        ax2.set_ylabel('MACD', fontsize=12)
        ax2.set_xlabel('Date', fontsize=12)
        ax2.legend(loc='best')
        ax2.grid(True, alpha=0.3)
        ax2.axhline(y=0, color='black', linestyle='-', linewidth=0.5)
        ax2.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))

        plt.tight_layout()

        # 保存图表
        output_path = os.path.join(output_dir, f'{symbol}_macd.png')
        plt.savefig(output_path, dpi=150, bbox_inches='tight')
        plt.close()

        return {
            "success": True,
            "chart_type": "macd",
            "file_path": output_path,
            "filename": f'{symbol}_macd.png',
            "symbol": symbol
        }

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": f"生成MACD图表失败: {str(e)}"}


def generate_rsi_chart_from_df(df, symbol, company_name, output_dir="."):
    """
    从 DataFrame 生成RSI指标图
    """
    try:
        # 创建图表
        fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(12, 10), gridspec_kw={'height_ratios': [2, 1]})

        # 绘制价格
        ax1.plot(df.index, df['Close'], label='Close Price', linewidth=2, color='#2E86AB')
        ax1.set_title(f'{company_name} ({symbol}) Price & RSI', fontsize=16, fontweight='bold')
        ax1.set_ylabel('Price (CNY)', fontsize=12)
        ax1.legend(loc='best')
        ax1.grid(True, alpha=0.3)
        ax1.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))

        # 绘制RSI
        if 'RSI' in df.columns:
            ax2.plot(df.index, df['RSI'], label='RSI', linewidth=2, color='#9C27B0')
            ax2.axhline(y=70, color='#D50000', linestyle='--', linewidth=1, label='Overbought (70)')
            ax2.axhline(y=30, color='#00C853', linestyle='--', linewidth=1, label='Oversold (30)')
            ax2.axhline(y=50, color='black', linestyle='-', linewidth=0.5, alpha=0.5)

            # 填充超买超卖区域
            ax2.fill_between(df.index, 70, 100, color='#D50000', alpha=0.1)
            ax2.fill_between(df.index, 0, 30, color='#00C853', alpha=0.1)

        ax2.set_ylabel('RSI', fontsize=12)
        ax2.set_xlabel('Date', fontsize=12)
        ax2.set_ylim(0, 100)
        ax2.legend(loc='best')
        ax2.grid(True, alpha=0.3)
        ax2.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))

        plt.tight_layout()

        # 保存图表
        output_path = os.path.join(output_dir, f'{symbol}_rsi.png')
        plt.savefig(output_path, dpi=150, bbox_inches='tight')
        plt.close()

        return {
            "success": True,
            "chart_type": "rsi",
            "file_path": output_path,
            "filename": f'{symbol}_rsi.png',
            "symbol": symbol
        }

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": f"生成RSI图表失败: {str(e)}"}


def generate_all_charts(symbol, market="cn", output_dir="."):
    """
    生成所有图表（主入口）
    """
    try:
        # 确保输出目录存在
        os.makedirs(output_dir, exist_ok=True)

        # 获取 TDX 数据
        print(f"正在获取 {symbol} 的数据...")
        stock_data = tdx_data_fetcher.get_stock_data(symbol, market, fetch_quote=True)
        if "error" in stock_data:
            return {"error": stock_data["error"]}

        company_name = stock_data.get("company_name", symbol)

        # 从 historical_data 重建 DataFrame
        hist_data = stock_data.get("historical_data", [])
        if not hist_data:
            return {"error": "没有历史数据可用"}

        df = pd.DataFrame(hist_data)
        df['Date'] = pd.to_datetime(df['date'])
        df = df.set_index('Date')
        df = df.sort_index()

        # 重命名为大写列名，保持一致
        df = df.rename(columns={
            'open': 'Open',
            'high': 'High',
            'low': 'Low',
            'close': 'Close',
            'volume': 'Volume'
        })

        # 计算技术指标（和 tdx_data_fetcher 保持一致）
        df['MA5'] = df['Close'].rolling(window=5).mean()
        df['MA10'] = df['Close'].rolling(window=10).mean()
        df['MA20'] = df['Close'].rolling(window=20).mean()
        df['MA60'] = df['Close'].rolling(window=60).mean()

        # MACD
        macd = ta.trend.MACD(df['Close'])
        df['MACD'] = macd.macd()
        df['MACD_Signal'] = macd.macd_signal()
        df['MACD_Diff'] = df['MACD'] - df['MACD_Signal']

        # RSI
        df['RSI'] = ta.momentum.RSIIndicator(df['Close']).rsi()

        results = {
            "symbol": symbol,
            "company_name": company_name,
            "market": market,
            "charts": [],
            "stock_data": stock_data
        }

        # 生成各个图表
        charts = [
            ('price_chart', generate_price_chart_from_df),
            ('candlestick', generate_candlestick_chart_from_df),
            ('macd', generate_macd_chart_from_df),
            ('rsi', generate_rsi_chart_from_df)
        ]

        for chart_name, chart_func in charts:
            print(f"正在生成 {chart_name} 图表...")
            result = chart_func(df, symbol, company_name, output_dir)
            if result.get('success'):
                results["charts"].append({
                    "type": chart_name,
                    "path": result["file_path"],
                    "filename": result["filename"]
                })
            else:
                results["charts"].append({
                    "type": chart_name,
                    "error": result.get("error")
                })
                print(f"  警告: {result.get('error')}")

        return results

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": f"生成综合仪表盘失败: {str(e)}"}


def main():
    parser = argparse.ArgumentParser(description='基于TDX数据生成股票图表')
    parser.add_argument('--symbol', type=str, required=True, help='股票代码')
    parser.add_argument('--market', type=str, choices=['cn'], default='cn', help='市场类型（仅支持cn）')
    parser.add_argument('--output-dir', type=str, default='.', help='输出目录')

    args = parser.parse_args()

    result = generate_all_charts(args.symbol, args.market, args.output_dir)
    print(json.dumps(result, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    import ta
    main()
