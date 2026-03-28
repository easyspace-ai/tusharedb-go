#!/usr/bin/env python3
import requests
import sys

url = "http://localhost:8787/api/tdx/search"
params = {"keyword": "000001"}

print(f"Testing: {url}")
print(f"Params: {params}")

try:
    response = requests.get(url, params=params, timeout=30)
    print(f"Status code: {response.status_code}")
    print(f"Headers: {dict(response.headers)}")
    print(f"Content: {response.text[:200]}")
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
    import traceback
    traceback.print_exc()
