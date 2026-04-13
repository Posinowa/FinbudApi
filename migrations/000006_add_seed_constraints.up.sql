-- Categories tablosuna unique constraint ekle
-- Default kategoriler için: aynı isim+type kombinasyonu tekrar edilemez
-- Kullanıcı kategorileri için: aynı user_id+name+type kombinasyonu tekrar edilemez

-- Default kategoriler için unique index (user_id NULL olanlar)
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_default_unique 
ON categories (name, type) 
WHERE is_default = true AND user_id IS NULL;

-- Kullanıcı kategorileri için unique index
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_user_unique 
ON categories (user_id, name, type) 
WHERE user_id IS NOT NULL;

-- Budgets tablosuna unique constraint ekle (aynı kullanıcı+kategori+ay+yıl tekrar edilemez)
CREATE UNIQUE INDEX IF NOT EXISTS idx_budgets_unique 
ON budgets (user_id, category_id, month, year);

-- Users tablosuna email unique constraint (zaten yoksa)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'users_email_key'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
    END IF;
END $$;