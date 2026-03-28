package service

import (
	"context"
	"fmt"

	"github.com/easyspace-ai/stock_api/internal/provider/stocksdk"
)

// FundService 资金服务
type FundService struct {
	client *stocksdk.Client
}

// NewFundService 创建资金服务
func NewFundService(client *stocksdk.Client) *FundService {
	return &FundService{client: client}
}

// GetFundFlow 获取资金流向
func (s *FundService) GetFundFlow(ctx context.Context, codes []string) ([]stocksdk.FundFlow, error) {
	if len(codes) == 0 {
		return nil, fmt.Errorf("codes cannot be empty")
	}
	return s.client.GetFundFlow(ctx, codes)
}

// GetPanelLargeOrder 获取盘口大单
func (s *FundService) GetPanelLargeOrder(ctx context.Context, codes []string) ([]stocksdk.PanelLargeOrder, error) {
	if len(codes) == 0 {
		return nil, fmt.Errorf("codes cannot be empty")
	}
	return s.client.GetPanelLargeOrder(ctx, codes)
}

// GetTradingCalendar 获取交易日历
func (s *FundService) GetTradingCalendar(ctx context.Context) ([]string, error) {
	return s.client.GetTradingCalendar(ctx)
}

// Search 搜索股票/板块
func (s *FundService) Search(ctx context.Context, keyword string) ([]stocksdk.SearchResult, error) {
	if keyword == "" {
		return nil, fmt.Errorf("keyword cannot be empty")
	}
	return s.client.Search(ctx, keyword)
}
