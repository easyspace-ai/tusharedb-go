#!/usr/bin/env python3
import os
import subprocess
import json

print("Environment:")
for k, v in os.environ.items():
    if 'proxy' in k.lower():
        print(f"  {k}={v}")

print("\nTesting curl directly...")
try:
    result = subprocess.run(
        ["curl", "-s", "http://localhost:8787/api/tdx/search?keyword=000001"],
        capture_output=True,
        text=True,
        timeout=10
    )
    print(f"Return code: {result.returncode}")
    print(f"Stdout: {result.stdout[:200]}")
    print(f"Stderr: {result.stderr}")
except Exception as e:
    print(f"Error: {e}")
