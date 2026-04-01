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
// GetAll retrieves all transactions for a user with filters
func (s *Service) GetAll(ctx context.Context, userID string, filter TransactionFilter) (*TransactionListResponse, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// Get transactions from repository
	transactions, total, err := s.repo.GetAll(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	// Build response with categories
	var responseData []TransactionResponse
	for _, t := range transactions {
		twc := TransactionWithCategory{Transaction: t}
		
		// Get category for each transaction
		cat, err := s.categoryRepo.GetByID(ctx, t.CategoryID)
		if err == nil {
			twc.Category = cat
		}
		
		responseData = append(responseData, ToTransactionResponse(&twc))
	}

	// Calculate total pages
	totalPages := total / filter.Limit
	if total%filter.Limit > 0 {
		totalPages++
	}

	return &TransactionListResponse{
		Data: responseData,
		Meta: PaginationMeta{
			Total:      total,
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

// Update updates a transaction
func (s *Service) Update(ctx context.Context, id string, userID string, req UpdateTransactionRequest) (*TransactionWithCategory, error) {
	// Get existing transaction
	result, err := s.repo.GetByIDWithCategory(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	// Check ownership
	if result.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update fields if provided
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, ErrInvalidAmount
		}
		result.Amount = *req.Amount
	}

	if req.CategoryID != nil {
		// Validate category exists and belongs to user
		cat, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, ErrCategoryNotFound
		}
		if cat.UserID != nil && *cat.UserID != userID {
			return nil, ErrCategoryNotFound
		}
		result.CategoryID = *req.CategoryID
	}

	if req.Description != nil {
		result.Description = req.Description
	}

	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return nil, ErrInvalidDate
		}
		result.Date = date
	}

	// Update timestamp
	result.UpdatedAt = time.Now()

	// Save to database
	if err := s.repo.Update(ctx, &result.Transaction); err != nil {
		return nil, err
	}

	// Get category for response
	cat, err := s.categoryRepo.GetByID(ctx, result.CategoryID)
	if err == nil {
		result.Category = cat
	}

	return result, nil
}


