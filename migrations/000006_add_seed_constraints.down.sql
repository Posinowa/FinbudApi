-- Eklenen constraint'leri kaldır
DROP INDEX IF EXISTS idx_categories_default_unique;
DROP INDEX IF EXISTS idx_categories_user_unique;
DROP INDEX IF EXISTS idx_budgets_unique;