#!/usr/bin/env bash
# 在仓库根目录执行：bash demo/run_all.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

demos=(
  realtime_quotes
  kline_minute
  market_universe
  capital_flow
  news_announcements
  global_assets
  market_special
  client_admin
  stockapi_sdk
  marketdata_rest
)

for d in "${demos[@]}"; do
  echo ""
  echo "========== demo/$d =========="
  go run "./demo/$d/" || echo "(跳过或失败: $d)"
done

echo ""
echo "========== demo/lake_query（需 CGO）=========="
if go run "./demo/lake_query/"; then
  :
else
  echo "lake_query 未执行成功：请安装 DuckDB CGO 依赖后单独运行 go run ./demo/lake_query/"
fi

echo ""
echo "全部示例已执行完毕。"
