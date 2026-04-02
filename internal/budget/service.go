package budget

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

// GetAll retrieves all budgets for a user in a specific month
func (s *Service) GetAll(ctx context.Context, userID string, monthStr string) (*BudgetListResponse, error) {
	// Parse month string (YYYY-MM) or use current month
	year, month, err := parseMonth(monthStr)
	if err != nil {
		return nil, ErrInvalidMonth
	}

	// Get budgets with spent amounts
	budgets, err := s.repo.GetAllByMonth(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	// Build response with categories
	var responseData []BudgetResponse
	for _, b := range budgets {
		// Get category
		cat, err := s.categoryRepo.GetByID(ctx, b.CategoryID)
		if err == nil {
			b.Category = cat
		}

		responseData = append(responseData, ToBudgetResponse(&b))
	}

	return &BudgetListResponse{
		Month: fmt.Sprintf("%d-%02d", year, month),
		Data:  responseData,
	}, nil
}

// parseMonth parses YYYY-MM format or returns current month if empty
func parseMonth(monthStr string) (int, int, error) {
	if monthStr == "" {
		now := time.Now()
		return now.Year(), int(now.Month()), nil
	}

	parts := strings.Split(monthStr, "-")
	if len(parts) != 2 {
		return 0, 0, ErrInvalidMonth
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 2000 || year > 2100 {
		return 0, 0, ErrInvalidMonth
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return 0, 0, ErrInvalidMonth
	}

	return year, month, nil
}

// Create creates a new budget
func (s *Service) Create(ctx context.Context, userID string, req CreateBudgetRequest) (*CreateBudgetResponse, error) {
	// Parse month
	year, month, err := parseMonth(req.Month)
	if err != nil {
		return nil, ErrInvalidMonth
	}

	// Check if category exists and belongs to user
	cat, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		return nil, ErrNotFound
	}
	if cat.UserID != nil && *cat.UserID != userID {
		return nil, ErrNotFound
	}

	// Check if budget already exists for this category and month
	exists, err := s.repo.Exists(ctx, userID, req.CategoryID, year, month)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyExists
	}

	// Create budget
	budget := &Budget{
		ID:         uuid.New().String(),
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount:     req.Limit,
		Month:      month,
		Year:       year,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save to database
	if err := s.repo.Create(ctx, budget); err != nil {
		return nil, err
	}

	response := ToCreateBudgetResponse(budget, cat)
	return &response, nil
}

// Update updates a budget (only limit can be changed)
func (s *Service) Update(ctx context.Context, id string, userID string, req UpdateBudgetRequest) (*CreateBudgetResponse, error) {
	// Get existing budget
	budget, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	// Check ownership
	if budget.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update limit
	budget.Amount = req.Limit
	budget.UpdatedAt = time.Now()

	// Save to database
	if err := s.repo.Update(ctx, budget); err != nil {
		return nil, err
	}

	// Get category for response
	cat, err := s.categoryRepo.GetByID(ctx, budget.CategoryID)
	if err != nil {
		cat = nil
	}

	response := ToCreateBudgetResponse(budget, cat)
	return &response, nil
}

// Delete deletes a budget
func (s *Service) Delete(ctx context.Context, id string, userID string) error {
	// Get existing budget
	budget, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotFound
	}

	// Check ownership
	if budget.UserID != userID {
		return ErrUnauthorized
	}

	// Delete from database
	return s.repo.Delete(ctx, id)
}
// GetByID retrieves a single budget by ID
func (s *Service) GetByID(ctx context.Context, budgetID string, userID string) (*BudgetResponse, error) {
	// Get budget from repository
	budget, err := s.repo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, ErrNotFound
	}

	// Check ownership
	if budget.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Get spent amount
	spent, err := s.repo.GetSpentAmount(ctx, userID, budget.CategoryID, budget.Year, budget.Month)
	if err != nil {
		spent = 0
	}

	// Get category
	cat, err := s.categoryRepo.GetByID(ctx, budget.CategoryID)
	if err != nil {
		cat = nil
	}

	// Build response
	budgetWithSpent := BudgetWithSpent{
		Budget:   *budget,
		Category: cat,
		Spent:    spent,
	}

	response := ToBudgetResponse(&budgetWithSpent)
	return &response, nil
}
