#!/usr/bin/env python3
import subprocess
import json
import pandas as pd
from datetime import datetime

result = subprocess.run(
    ["curl", "-s", "http://localhost:8787/api/tdx/kline-all/ths?code=000001&type=day"],
    capture_output=True,
    text=True,
    timeout=30
)

data = json.loads(result.stdout)
print("Keys in data:", list(data.keys()))

if 'data' in data and 'list' in data['data']:
    klines = data['data']['list']
    print(f"Got {len(klines)} K-line entries")
    print("\nFirst 3 entries:")
    for k in klines[:3]:
        print(k)

    print("\nLast 3 entries:")
    for k in klines[-3:]:
        print(k)

    # Try to create DataFrame
    df = pd.DataFrame(klines)
    print("\nDataFrame columns:", list(df.columns))
    print("\nFirst 3 rows:")
    print(df.head(3))

    # Try to parse datetime
    if 'Time' in df.columns:
        print("\nTime column sample:", df['Time'].head(5))

        # Try different parsing methods
        print("\nTrying pd.to_datetime with format...")
        try:
            # 直接处理，不使用 dt accessor
            dates = []
            for t in df['Time']:
                s = str(t)
                if 'T' in s:
                    date_part = s.split('T')[0]
                else:
                    date_part = s[:10]
                dates.append(date_part)
            df['Date'] = pd.to_datetime(dates)
            print("Success! Dates:")
            print(df['Date'].head())
        except Exception as e:
            print(f"Error: {e}")
            import traceback
            traceback.print_exc()
