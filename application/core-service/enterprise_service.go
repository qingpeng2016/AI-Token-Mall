package coreservice

import (
	"context"
	"strings"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/request"
	"github.com/qingpeng2016/ai-token-mall/domain/rest/response"
)

type EnterpriseService struct {
	enterpriseRepo repository.EnterpriseRepo
}

func NewEnterpriseService(enterpriseRepo repository.EnterpriseRepo) *EnterpriseService {
	return &EnterpriseService{enterpriseRepo: enterpriseRepo}
}

func (s *EnterpriseService) SubmitInquiry(ctx context.Context, req *request.SubmitEnterpriseInquiryReq) (uint, error) {
	row := &entity.EnterpriseInquiry{
		CompanyName: strings.TrimSpace(req.CompanyName),
		ContactName: strings.TrimSpace(req.ContactName),
		Phone:       strings.TrimSpace(req.Phone),
		Status:      "pending",
	}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		row.Email = &email
	}
	if err := s.enterpriseRepo.CreateInquiry(ctx, row); err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (s *EnterpriseService) ListProducts(ctx context.Context) (*response.EnterpriseProductListResp, error) {
	rows, err := s.enterpriseRepo.ListActiveProducts(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]response.EnterpriseProductItemResp, 0, len(rows))
	for i := range rows {
		p := &rows[i]
		features := decodeStringJSONArray(p.Features)
		if features == nil {
			features = []string{}
		}
		items = append(items, response.EnterpriseProductItemResp{
			Code:        p.Code,
			Name:        p.Name,
			Badge:       p.Badge,
			PriceHint:   p.PriceHint,
			Seats:       p.Seats,
			Features:    features,
			Tagline:     p.Tagline,
			ButtonLabel: p.ButtonLabel,
			IsFeatured:  p.IsFeatured != 0,
			Sort:        p.Sort,
		})
	}
	return &response.EnterpriseProductListResp{Products: items}, nil
}
