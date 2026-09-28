package repository

import (
	"context"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
)

type TutorialRepo interface {
	ListActiveCategories(ctx context.Context) ([]entity.TutorialCategory, error)
	ListPublishedArticles(ctx context.Context, categoryCode string) ([]entity.TutorialArticle, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*entity.TutorialArticle, error)
	GetCategoryByID(ctx context.Context, id uint) (*entity.TutorialCategory, error)
}
