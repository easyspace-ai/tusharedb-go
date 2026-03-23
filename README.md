# tusharedb-go

New Go implementation of TushareDB.

Current planning documents:

- `docs/V1_API_DRAFT.md`: public v1 library API draft
- `docs/BOOTSTRAP_CHECKLIST.md`: scaffold and implementation checklist

Goals:

- Library-first design
- Parquet as source-of-truth storage
- DuckDB as query engine
- Optimized for daily backtesting and stock screening
- Easy dataset expansion over time
