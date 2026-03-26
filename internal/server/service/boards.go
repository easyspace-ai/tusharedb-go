package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/easyspace-ai/tusharedb-go/internal/provider/stocksdk"
)

// BoardService 板块服务
type BoardService struct {
	client *stocksdk.Client
}

// NewBoardService 创建板块服务
func NewBoardService(client *stocksdk.Client) *BoardService {
	return &BoardService{client: client}
}

// GetIndustryList 获取行业板块列表
func (s *BoardService) GetIndustryList(ctx context.Context) ([]stocksdk.IndustryBoard, error) {
	return s.client.GetIndustryList(ctx)
}

// GetConceptList 获取概念板块列表
func (s *BoardService) GetConceptList(ctx context.Context) ([]stocksdk.ConceptBoard, error) {
	return s.client.GetConceptList(ctx)
}

// GetIndustryConstituents 获取行业板块成分股
func (s *BoardService) GetIndustryConstituents(ctx context.Context, code string) ([]stocksdk.IndustryBoardConstituent, error) {
	return s.client.GetIndustryConstituents(ctx, code)
}

// GetConceptConstituents 获取概念板块成分股
func (s *BoardService) GetConceptConstituents(ctx context.Context, code string) ([]stocksdk.ConceptBoardConstituent, error) {
	return s.client.GetConceptConstituents(ctx, code)
}

// GetIndustryKline 获取行业板块K线
func (s *BoardService) GetIndustryKline(ctx context.Context, code, period string) ([]stocksdk.IndustryBoardKline, error) {
	var klinePeriod stocksdk.KlinePeriod
	switch strings.ToLower(period) {
	case "weekly", "week":
		klinePeriod = stocksdk.KlinePeriodWeekly
	case "monthly", "month":
		klinePeriod = stocksdk.KlinePeriodMonthly
	default:
		klinePeriod = stocksdk.KlinePeriodDaily
	}
	
	return s.client.GetIndustryKline(ctx, code, &stocksdk.BoardKlineOptions{
		Period: klinePeriod,
	})
}

// GetConceptKline 获取概念板块K线
func (s *BoardService) GetConceptKline(ctx context.Context, code, period string) ([]stocksdk.ConceptBoardKline, error) {
	var klinePeriod stocksdk.KlinePeriod
	switch strings.ToLower(period) {
	case "weekly", "week":
		klinePeriod = stocksdk.KlinePeriodWeekly
	case "monthly", "month":
		klinePeriod = stocksdk.KlinePeriodMonthly
	default:
		klinePeriod = stocksdk.KlinePeriodDaily
	}
	
	return s.client.GetConceptKline(ctx, code, &stocksdk.BoardKlineOptions{
		Period: klinePeriod,
	})
}

// GetIndustrySpot 获取行业板块实时指标
func (s *BoardService) GetIndustrySpot(ctx context.Context, code string) ([]stocksdk.IndustryBoardSpot, error) {
	return s.client.GetIndustrySpot(ctx, code)
}

// GetConceptSpot 获取概念板块实时指标
func (s *BoardService) GetConceptSpot(ctx context.Context, code string) ([]stocksdk.ConceptBoardSpot, error) {
	return s.client.GetConceptSpot(ctx, code)
}

// SearchBoard 搜索板块（根据名称查找代码）
func (s *BoardService) SearchBoard(ctx context.Context, keyword string) ([]stocksdk.IndustryBoard, error) {
	// 获取所有板块列表
	industries, err := s.client.GetIndustryList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get industry list: %w", err)
	}
	
	concepts, err := s.client.GetConceptList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get concept list: %w", err)
	}
	
	// 合并并过滤
	keyword = strings.ToLower(keyword)
	var results []stocksdk.IndustryBoard
	
	for _, board := range industries {
		if strings.Contains(strings.ToLower(board.Name), keyword) ||
			strings.Contains(strings.ToLower(board.Code), keyword) {
			results = append(results, board)
		}
	}
	
	for _, board := range concepts {
		if strings.Contains(strings.ToLower(board.Name), keyword) ||
			strings.Contains(strings.ToLower(board.Code), keyword) {
			results = append(results, board)
		}
	}
	
	return results, nil
}
