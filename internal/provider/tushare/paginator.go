package tushare

import (
	"context"
	"fmt"

	"github.com/easyspace-ai/stock_api/internal/provider"
)

func (c *Client) FetchAllPages(ctx context.Context, req provider.Request) (*provider.Response, error) {
	limit := req.Page.Limit
	if limit <= 0 {
		limit = c.cfg.PageLimit
	}

	offset := req.Page.Offset
	combined := &provider.Response{}

	for {
		pageReq := req
		pageReq.Page = provider.PageRequest{
			Limit:  limit,
			Offset: offset,
		}

		pageResp, err := c.fetchOne(ctx, pageReq)
		if err != nil {
			return nil, fmt.Errorf("fetch page offset=%d: %w", offset, err)
		}

		if len(combined.Fields) == 0 {
			combined.Fields = pageResp.Fields
		}
		combined.Rows = append(combined.Rows, pageResp.Rows...)

		if len(pageResp.Rows) < limit {
			break
		}

		offset += limit
	}

	return combined, nil
}
