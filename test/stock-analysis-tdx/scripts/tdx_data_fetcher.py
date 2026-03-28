#!/usr/bin/env python3
"""
TDX通达信股票市场数据获取脚本

功能：
- 通过 TDX API 获取实时股票行情数据
- 计算技术指标（MA、MACD、RSI）
- 获取历史价格数据（支持前复权）
- 仅支持A股市场

依赖：
- requests: HTTP请求库
- ta: 技术分析库
- pandas: 数据处理
"""

import argparse
import json
import os
import sys
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, List

# 禁用代理
os.environ['NO_PROXY'] = '*'
os.environ['no_proxy'] = '*'
os.environ.pop('HTTP_PROXY', None)
os.environ.pop('http_proxy', None)
os.environ.pop('HTTPS_PROXY', None)
os.environ.pop('https_proxy', None)
os.environ.pop('ALL_PROXY', None)
os.environ.pop('all_proxy', None)

import numpy as np
import pandas as pd
import ta

# TDX API 配置
DEFAULT_TDX_API_BASE = os.getenv("TDX_API_BASE_URL", "http://localhost:8787/api/tdx")


def _convert_price(li: int) -> float:
    """将TDX的厘单位转换为元，处理负数情况（前复权数据）"""
    if li is None:
        return 0.0
    # 前复权数据可能是负数，表示相对价格
    # 取绝对值转换
    return round(abs(li) / 1000.0, 2)


def _convert_volume(volume: int) -> int:
    """将TDX的手单位转换为股"""
    return volume * 100 if volume else 0


def _call_tdx_api(endpoint: str, params: Optional[Dict] = None, method: str = "GET") -> Optional[Dict]:
    """
    调用 TDX API (使用 curl 避免代理问题)

    参数:
        endpoint: API 端点（如 /quote）
        params: 查询参数
        method: HTTP 方法

    返回:
        API 响应数据，失败返回 None
    """
    import subprocess
    import urllib.parse

    url = f"{DEFAULT_TDX_API_BASE.rstrip('/')}{endpoint}"

    if params:
        query_string = urllib.parse.urlencode(params)
        url = f"{url}?{query_string}"

    try:
        result = subprocess.run(
            ["curl", "-s", url],
            capture_output=True,
            text=True,
            timeout=30
        )

        if result.returncode != 0:
            print(f"curl 调用失败: {result.stderr}", file=sys.stderr)
            return None

        response_data = json.loads(result.stdout)

        if response_data.get("code") != 0:
            print(f"TDX API 返回错误: {response_data.get('message')}", file=sys.stderr)
            return None

        return response_data.get("data")

    except subprocess.TimeoutExpired:
        print("TDX API 请求超时", file=sys.stderr)
        return None
    except json.JSONDecodeError as e:
        print(f"解析 TDX API 响应失败: {e}", file=sys.stderr)
        print(f"响应内容: {result.stdout[:200] if 'result' in locals() else 'N/A'}", file=sys.stderr)
        return None
    except Exception as e:
        print(f"调用 TDX API 失败: {e}", file=sys.stderr)
        return None


def get_stock_name(code: str) -> Optional[str]:
    """
    通过搜索获取股票名称

    参数:
        code: 股票代码

    返回:
        股票名称，失败返回 None
    """
    # 只取6位数字代码
    clean_code = ''.join([c for c in code if c.isdigit()])
    if len(clean_code) != 6:
        return None

    data = _call_tdx_api("/search", {"keyword": clean_code})
    if data and isinstance(data, list):
        for item in data:
            if item.get("code") == clean_code:
                return item.get("name")
    return None


def get_stock_quote(code: str) -> Optional[Dict]:
    """
    获取股票五档行情

    参数:
        code: 股票代码

    返回:
        行情数据字典
    """
    data = _call_tdx_api("/quote", {"code": code})
    if data and isinstance(data, list) and len(data) > 0:
        return data[0]
    return None


def get_kline_data(code: str, kline_type: str = "day") -> Optional[pd.DataFrame]:
    """
    获取K线数据（/api/tdx/kline 接口已包含前复权）

    参数:
        code: 股票代码
        kline_type: K线类型 (minute1/minute5/minute15/minute30/hour/day/week/month)

    返回:
        DataFrame 格式的K线数据
    """
    clean_code = ''.join([c for c in code if c.isdigit()])

    # 使用标准K线接口（已包含前复权，更稳定）
    data = _call_tdx_api("/kline", {"code": clean_code, "type": kline_type})
    if data:
        # 兼容两种返回格式："List" 或 "list"
        klines = data.get("List") or data.get("list")
        if klines:
            df = pd.DataFrame(klines)
            return _process_kline_df(df)

    return None


def _process_kline_df(df: pd.DataFrame) -> pd.DataFrame:
    """处理K线DataFrame，统一格式"""
    # 统一列名
    col_mapping = {
        "Time": "datetime",
        "time": "datetime",
        "Open": "Open",
        "open": "Open",
        "High": "High",
        "high": "High",
        "Low": "Low",
        "low": "Low",
        "Close": "Close",
        "close": "Close",
        "Volume": "Volume",
        "volume": "Volume",
        "Amount": "Amount",
        "amount": "Amount",
        "Last": "Last",
        "last": "Last",
    }

    df = df.rename(columns={k: v for k, v in col_mapping.items() if k in df.columns})

    # 确保必要列存在
    required_cols = ["Open", "High", "Low", "Close", "Volume"]
    for col in required_cols:
        if col not in df.columns:
            df[col] = 0

    # 转换价格单位（厘 -> 元）
    for col in ["Open", "High", "Low", "Close", "Last", "Amount"]:
        if col in df.columns:
            df[col] = pd.to_numeric(df[col], errors="coerce").apply(lambda x: _convert_price(int(x)) if pd.notna(x) else x)

    # 转换成交量（手 -> 股）
    if "Volume" in df.columns:
        df["Volume"] = pd.to_numeric(df["Volume"], errors="coerce").apply(lambda x: _convert_volume(int(x)) if pd.notna(x) else x)

    # 处理时间列
    dates = []
    if "datetime" in df.columns:
        for dt_str in df["datetime"]:
            if pd.isna(dt_str):
                dates.append(None)
            else:
                # 手动解析日期时间字符串，避免时区问题
                s = str(dt_str)
                # 只取 YYYY-MM-DD 部分
                if 'T' in s:
                    date_part = s.split('T')[0]
                else:
                    date_part = s[:10]
                dates.append(date_part)
        df["Date"] = pd.to_datetime(dates)
    else:
        df["Date"] = pd.date_range(end=datetime.now(), periods=len(df), freq="D")

    df = df.sort_values("Date").set_index("Date")

    # 确保数据类型正确
    for col in ["Open", "High", "Low", "Close", "Volume"]:
        df[col] = pd.to_numeric(df[col], errors="coerce")

    return df.dropna(subset=["Close"])


def get_stock_data(
    symbol: str,
    market: str = "cn",
    period: str = "1y",
    fetch_quote: bool = True
) -> Dict[str, Any]:
    """
    获取股票数据（主入口函数）

    参数:
        symbol: 股票代码（如 000001, 600519）
        market: 市场类型（仅支持 cn）
        period: 时间周期（1y=1年, 6mo=6个月, 3mo=3个月）
        fetch_quote: 是否获取实时行情

    返回:
        dict: 包含行情数据、技术指标的字典
    """
    try:
        if market != "cn":
            return {"error": "TDX数据源仅支持A股市场"}

        # 清理股票代码（只保留数字）
        clean_code = ''.join([c for c in symbol if c.isdigit()])
        if len(clean_code) != 6:
            return {"error": f"无效的股票代码格式: {symbol}，应为6位数字"}

        # 获取股票名称
        company_name = get_stock_name(clean_code) or "N/A"

        # 获取K线数据（/api/tdx/kline 已包含前复权）
        df = get_kline_data(clean_code, "day")
        if df is None or df.empty:
            return {"error": f"无法获取股票 {clean_code} 的K线数据，请检查TDX API服务是否正常"}

        # 根据period过滤数据
        days_map = {"1y": 365, "6mo": 180, "3mo": 90, "1mo": 30}
        days = days_map.get(period, 365)
        cutoff_date = datetime.now() - timedelta(days=days + 30)
        df = df[df.index >= cutoff_date].copy()

        if df.empty:
            return {"error": f"股票 {clean_code} 的历史数据不足"}

        # 计算技术指标
        df = df.copy()

        # 移动平均线
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

        # 布林带
        bb = ta.volatility.BollingerBands(df['Close'])
        df['BB_Upper'] = bb.bollinger_hband()
        df['BB_Middle'] = bb.bollinger_mavg()
        df['BB_Lower'] = bb.bollinger_lband()

        # 成交量移动平均
        df['Volume_MA5'] = df['Volume'].rolling(window=5).mean()
        df['Volume_MA10'] = df['Volume'].rolling(window=10).mean()

        # 获取最新数据
        latest = df.iloc[-1]
        prev = df.iloc[-2] if len(df) > 1 else latest

        # 获取实时行情
        current_price = round(latest['Close'], 2)
        price_change = 0.0
        price_change_pct = 0.0
        volume = int(latest['Volume']) if pd.notna(latest['Volume']) else 0

        if fetch_quote:
            quote = get_stock_quote(clean_code)
            if quote:
                k_data = quote.get('K', {})
                if k_data:
                    # TDX返回价格单位是厘
                    last_close = _convert_price(k_data.get('Last', 0))
                    current_close = _convert_price(k_data.get('Close', 0))
                    if last_close > 0:
                        current_price = current_close
                        price_change = current_close - last_close
                        price_change_pct = (price_change / last_close) * 100
                volume = quote.get('TotalHand', 0) * 100  # 手->股

        # 计算52周高低点（如果数据足够）
        high_52w = df['High'].max() if len(df) >= 30 else 0
        low_52w = df['Low'].min() if len(df) >= 30 else 0

        # 构建返回结果
        result = {
            "symbol": clean_code,
            "market": market,
            "company_name": company_name,
            "current_price": float(round(current_price, 2)),
            "price_change": float(round(price_change, 2)),
            "price_change_pct": float(round(price_change_pct, 2)),
            "volume": int(volume),
            "high_52w": float(round(high_52w, 2)) if high_52w else 0.0,
            "low_52w": float(round(low_52w, 2)) if low_52w else 0.0,
            "market_cap": 0,  # TDX不直接提供市值
            "pe_ratio": 0,   # TDX不直接提供PE
            "pb_ratio": 0,   # TDX不直接提供PB

            # 技术指标
            "technical_indicators": {
                "MA5": float(round(latest['MA5'], 2)) if pd.notna(latest['MA5']) else None,
                "MA10": float(round(latest['MA10'], 2)) if pd.notna(latest['MA10']) else None,
                "MA20": float(round(latest['MA20'], 2)) if pd.notna(latest['MA20']) else None,
                "MA60": float(round(latest['MA60'], 2)) if pd.notna(latest['MA60']) else None,
                "MACD": float(round(latest['MACD'], 4)) if pd.notna(latest['MACD']) else None,
                "MACD_Signal": float(round(latest['MACD_Signal'], 4)) if pd.notna(latest['MACD_Signal']) else None,
                "MACD_Diff": float(round(latest['MACD_Diff'], 4)) if pd.notna(latest['MACD_Diff']) else None,
                "RSI": float(round(latest['RSI'], 2)) if pd.notna(latest['RSI']) else None,
                "BB_Upper": float(round(latest['BB_Upper'], 2)) if pd.notna(latest['BB_Upper']) else None,
                "BB_Middle": float(round(latest['BB_Middle'], 2)) if pd.notna(latest['BB_Middle']) else None,
                "BB_Lower": float(round(latest['BB_Lower'], 2)) if pd.notna(latest['BB_Lower']) else None,
            },

            # 趋势判断
            "trend_analysis": {
                "multi_bullish": bool(check_bullish_alignment(df)),
                "ma_signal": get_ma_signal(latest),
                "macd_signal": "金叉" if pd.notna(latest['MACD_Diff']) and latest['MACD_Diff'] > 0 else "死叉" if pd.notna(latest['MACD_Diff']) else "信号不明",
                "rsi_signal": get_rsi_signal(latest['RSI']) if pd.notna(latest['RSI']) else "信号不明",
            },

            # 历史价格数据（最近30天）
            "historical_data": [
                {
                    "date": idx.strftime('%Y-%m-%d'),
                    "open": float(round(row['Open'], 2)) if pd.notna(row['Open']) else 0.0,
                    "high": float(round(row['High'], 2)) if pd.notna(row['High']) else 0.0,
                    "low": float(round(row['Low'], 2)) if pd.notna(row['Low']) else 0.0,
                    "close": float(round(row['Close'], 2)) if pd.notna(row['Close']) else 0.0,
                    "volume": int(row['Volume']) if pd.notna(row['Volume']) else 0
                }
                for idx, row in df.tail(30).iterrows()
            ],

            "data_source": "TDX(通达信)",
            "data_timestamp": datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        }

        return result

    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": f"获取数据失败: {str(e)}"}


def check_bullish_alignment(df: pd.DataFrame) -> bool:
    """
    检查是否形成多头排列

    多头排列定义: MA5 > MA10 > MA20 > MA60
    """
    latest = df.iloc[-1]
    try:
        if (pd.notna(latest['MA5']) and pd.notna(latest['MA10']) and
            pd.notna(latest['MA20']) and pd.notna(latest['MA60'])):
            return (latest['MA5'] > latest['MA10'] > latest['MA20'] > latest['MA60'])
    except:
        pass
    return False


def get_ma_signal(latest: pd.Series) -> str:
    """获取均线信号"""
    current = latest['Close']
    if pd.notna(latest['MA5']):
        if current > latest['MA5'] > latest.get('MA10', 0):
            return "强势"
        elif current < latest['MA5'] < latest.get('MA10', 0):
            return "弱势"
        elif current > latest['MA5']:
            return "反弹"
        else:
            return "调整"
    return "信号不明"


def get_rsi_signal(rsi: float) -> str:
    """获取RSI信号"""
    if rsi > 70:
        return "超买"
    elif rsi < 30:
        return "超卖"
    elif rsi > 50:
        return "强势"
    else:
        return "弱势"


def get_index_kline_data(code: str, kline_type: str = "day") -> Optional[pd.DataFrame]:
    """
    获取指数K线数据

    参数:
        code: 指数代码（如 sh000001 表示上证指数）
        kline_type: K线类型

    返回:
        DataFrame 格式的K线数据
    """
    # 使用指数接口
    data = _call_tdx_api("/index", {"code": code, "type": kline_type})
    if data:
        klines = data.get("List") or data.get("list")
        if klines:
            df = pd.DataFrame(klines)
            return _process_kline_df(df)
    return None


def get_market_index(market: str = "cn") -> Dict[str, Any]:
    """
    获取市场指数数据

    参数:
        market: 市场类型

    返回:
        包含指数数据的字典
    """
    try:
        if market != "cn":
            return {"error": "TDX数据源仅支持A股市场"}

        def fetch_index_data(code: str, name: str) -> Optional[Dict]:
            """内部函数：获取单个指数数据"""
            df = get_index_kline_data(code, "day")
            if df is None or df.empty:
                return None

            # 计算技术指标
            df = df.copy()
            df['MA5'] = df['Close'].rolling(window=5).mean()
            df['MA10'] = df['Close'].rolling(window=10).mean()
            df['MA20'] = df['Close'].rolling(window=20).mean()

            latest = df.iloc[-1]
            prev = df.iloc[-2] if len(df) > 1 else latest

            price_change = latest['Close'] - prev['Close']
            price_change_pct = (price_change / prev['Close']) * 100 if prev['Close'] > 0 else 0

            return {
                "symbol": code,
                "market": "cn",
                "company_name": name,
                "current_price": float(round(latest['Close'], 2)),
                "price_change": float(round(price_change, 2)),
                "price_change_pct": float(round(price_change_pct, 2)),
                "volume": int(latest['Volume']) if pd.notna(latest['Volume']) else 0,
                "high_52w": float(round(df['High'].max(), 2)) if len(df) >= 30 else 0.0,
                "low_52w": float(round(df['Low'].min(), 2)) if len(df) >= 30 else 0.0,
                "market_cap": 0,
                "pe_ratio": 0,
                "pb_ratio": 0,
                "technical_indicators": {
                    "MA5": float(round(latest['MA5'], 2)) if pd.notna(latest['MA5']) else None,
                    "MA10": float(round(latest['MA10'], 2)) if pd.notna(latest['MA10']) else None,
                    "MA20": float(round(latest['MA20'], 2)) if pd.notna(latest['MA20']) else None,
                },
                "trend_analysis": {},
                "historical_data": [
                    {
                        "date": idx.strftime('%Y-%m-%d'),
                        "open": float(round(row['Open'], 2)) if pd.notna(row['Open']) else 0.0,
                        "high": float(round(row['High'], 2)) if pd.notna(row['High']) else 0.0,
                        "low": float(round(row['Low'], 2)) if pd.notna(row['Low']) else 0.0,
                        "close": float(round(row['Close'], 2)) if pd.notna(row['Close']) else 0.0,
                        "volume": int(row['Volume']) if pd.notna(row['Volume']) else 0
                    }
                    for idx, row in df.tail(30).iterrows()
                ],
                "data_source": "TDX(通达信)",
                "data_timestamp": datetime.now().strftime('%Y-%m-%d %H:%M:%S')
            }

        # 上证指数
        sh_data = fetch_index_data("sh000001", "上证指数")

        # 深证成指
        sz_data = fetch_index_data("sz399001", "深证成指")

        return {
            "cn_market": {
                "sh_index": sh_data,
                "sz_index": sz_data,
            }
        }
    except Exception as e:
        import traceback
        traceback.print_exc()
        return {"error": f"获取市场指数失败: {str(e)}"}


def main():
    parser = argparse.ArgumentParser(description='通过TDX API获取股票市场数据')
    parser.add_argument('--symbol', type=str, help='股票代码（如 000001 或 600519）')
    parser.add_argument('--market', type=str, choices=['cn'], default='cn', help='市场类型（仅支持cn=A股）')
    parser.add_argument('--mode', type=str, choices=['stock', 'market', 'both'], default='stock',
                       help='获取模式（stock=个股, market=大盘, both=两者）')
    parser.add_argument('--period', type=str, default='1y', help='数据周期（1y, 6mo, 3mo）')
    parser.add_argument(
        '--no-quote',
        action='store_true',
        help='不请求实时行情数据',
    )

    args = parser.parse_args()

    # 如果获取大盘数据
    if args.mode in ['market', 'both']:
        market_data = get_market_index(args.market)
        print(json.dumps(market_data, indent=2, ensure_ascii=False))

    # 如果获取个股数据
    if args.mode in ['stock', 'both']:
        if not args.symbol:
            print(json.dumps({"error": "个股模式需要提供 --symbol 参数"}, indent=2, ensure_ascii=False))
            sys.exit(1)

        stock_data = get_stock_data(
            args.symbol,
            args.market,
            args.period,
            fetch_quote=not args.no_quote,
        )
        print(json.dumps(stock_data, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()
