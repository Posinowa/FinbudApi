CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(50),
    type VARCHAR(10) CHECK (type IN ('income', 'expense')) NOT NULL,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Varsayılan kategorileri ekle
INSERT INTO categories (name, icon, type, is_default) VALUES
('Maaş', '💰', 'income', true),
('Freelance', '💻', 'income', true),
('Yatırım', '📈', 'income', true),
('Diğer Gelir', '💵', 'income', true),
('Yemek', '🍔', 'expense', true),
('Ulaşım', '🚗', 'expense', true),
('Alışveriş', '🛒', 'expense', true),
('Faturalar', '📄', 'expense', true),
('Sağlık', '🏥', 'expense', true),
('Eğlence', '🎬', 'expense', true),
('Diğer Gider', '📦', 'expense', true);