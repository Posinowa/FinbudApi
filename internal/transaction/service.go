package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Posinowa/FinbudApp/internal/category"
)

type Service struct {
	repo         *Repository
	categoryRepo *category.Repository
}

func NewService(repo *Repository, categoryRepo *category.Repository) *Service {
	return &Service{
		repo:         repo,
		categoryRepo: categoryRepo,
	}
}

// Create creates a new transaction
func (s *Service) Create(ctx context.Context, input CreateTransactionInput) (*TransactionWithCategory, error) {
	// Validate transaction type
	if input.Type != TypeIncome && input.Type != TypeExpense {
		return nil, ErrInvalidType
	}

	// Validate amount
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Parse and validate date
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}

	// Check if category exists and belongs to user
	cat, err := s.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		return nil, ErrCategoryNotFound
	}
	if cat.UserID != nil && *cat.UserID != input.UserID {
		return nil, ErrCategoryNotFound
	}

	// Create transaction
	transaction := &Transaction{
		ID:          uuid.New().String(),
		UserID:      input.UserID,
		CategoryID:  input.CategoryID,
		Amount:      input.Amount,
		Type:        input.Type,
		Date:        date,
		Description: input.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.repo.Create(ctx, transaction); err != nil {
		return nil, err
	}

	// Return with category
	return &TransactionWithCategory{
		Transaction: *transaction,
		Category:    cat,
	}, nil
}
// GetByID retrieves a transaction by ID with category
func (s *Service) GetByID(ctx context.Context, id string, userID string) (*TransactionWithCategory, error) {
	// Get transaction
	result, err := s.repo.GetByIDWithCategory(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	// Check ownership
	if result.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Get category
	cat, err := s.categoryRepo.GetByID(ctx, result.CategoryID)
	if err == nil {
		result.Category = cat
	}

	return result, nil
}
