package category

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
)

var (
	ErrCategoryNotFound    = errors.New("kategori bulunamadi")
	ErrForbidden           = errors.New("bu islemi yapmaya yetkiniz yok")
	ErrDefaultCategory     = errors.New("varsayilan kategoriler silinemez")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context, userID string, categoryType *string) ([]Category, int, error) {
	categories, err := s.repo.GetAll(ctx, userID, categoryType)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if categories == nil {
		categories = []Category{}
	}

	return categories, http.StatusOK, nil
}

func (s *Service) Create(ctx context.Context, userID string, req CreateCategoryRequest) (*Category, int, error) {
	category, err := s.repo.Create(ctx, userID, req.Name, req.Icon, req.Type)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return category, http.StatusCreated, nil
}

func (s *Service) Update(ctx context.Context, userID, categoryID string, req UpdateCategoryRequest) (*Category, int, error) {
	// Kategoriyi bul
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, http.StatusNotFound, ErrCategoryNotFound
		}
		return nil, http.StatusInternalServerError, err
	}

	// Varsayılan kategori güncellenemez
	if category.IsDefault {
		return nil, http.StatusForbidden, ErrForbidden
	}

	// Kullanıcı sadece kendi kategorisini güncelleyebilir
	if category.UserID == nil || *category.UserID != userID {
		return nil, http.StatusForbidden, ErrForbidden
	}

	// Güncelle
	updated, err := s.repo.Update(ctx, categoryID, req.Name, req.Icon, req.Type)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return updated, http.StatusOK, nil
}

func (s *Service) Delete(ctx context.Context, userID, categoryID string) (int, error) {
	// Kategoriyi bul
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return http.StatusNotFound, ErrCategoryNotFound
		}
		return http.StatusInternalServerError, err
	}

	// Varsayılan kategori silinemez
	if category.IsDefault {
		return http.StatusBadRequest, ErrDefaultCategory
	}

	// Kullanıcı sadece kendi kategorisini silebilir
	if category.UserID == nil || *category.UserID != userID {
		return http.StatusForbidden, ErrForbidden
	}

	// Sil
	err = s.repo.Delete(ctx, categoryID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusNoContent, nil
}
// GetByID retrieves a single category by ID
func (s *Service) GetByID(ctx context.Context, userID string, categoryID string) (*Category, int, error) {
	category, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, http.StatusNotFound, errors.New("category not found")
	}

	// Check access: default categories are accessible by all, custom categories only by owner
	if category.UserID != nil && *category.UserID != userID {
		return nil, http.StatusForbidden, errors.New("access denied")
	}

	return category, http.StatusOK, nil
}