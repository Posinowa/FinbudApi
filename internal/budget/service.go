package budget

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
