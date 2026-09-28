package mysql

import (
	"context"
	"errors"

	"github.com/qingpeng2016/ai-token-mall/domain/persistent/entity"
	"github.com/qingpeng2016/ai-token-mall/domain/persistent/repository"
	"gorm.io/gorm"
)

type TutorialImpl struct {
	db *gorm.DB
}

func NewTutorialImpl(db *gorm.DB) repository.TutorialRepo {
	return &TutorialImpl{db: db}
}

func (r *TutorialImpl) ListActiveCategories(ctx context.Context) ([]entity.TutorialCategory, error) {
	var rows []entity.TutorialCategory
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("sort ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *TutorialImpl) ListPublishedArticles(ctx context.Context, categoryCode string) ([]entity.TutorialArticle, error) {
	q := r.db.WithContext(ctx).
		Table("tutorial_article AS a").
		Select("a.*").
		Where("a.status = ?", "published")
	if categoryCode != "" {
		q = q.Joins("INNER JOIN tutorial_category c ON c.id = a.category_id AND c.status = ?", "active").
			Where("c.code = ?", categoryCode)
	}
	var rows []entity.TutorialArticle
	err := q.Order("a.published_at DESC, a.sort ASC, a.id DESC").Find(&rows).Error
	return rows, err
}

func (r *TutorialImpl) FindPublishedBySlug(ctx context.Context, slug string) (*entity.TutorialArticle, error) {
	var row entity.TutorialArticle
	err := r.db.WithContext(ctx).
		Where("slug = ? AND status = ?", slug, "published").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TutorialImpl) GetCategoryByID(ctx context.Context, id uint) (*entity.TutorialCategory, error) {
	var row entity.TutorialCategory
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
