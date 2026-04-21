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
('Maaş', 'assets/icons/para.png', 'income', true),
('Freelance', 'assets/icons/laptop.png', 'income', true),
('Yatırım', 'assets/icons/para_akis.png', 'income', true),
('Diğer Gelir', 'assets/icons/gelir_cüzdan.png', 'income', true),
('Yemek', 'assets/icons/restorant.png', 'expense', true),
('Ulaşım', 'assets/icons/araba.png', 'expense', true),
('Alışveriş', 'assets/icons/market_arabasi.png', 'expense', true),
('Faturalar', 'assets/icons/fatura.png', 'expense', true),
('Sağlık', 'assets/icons/saglik.png', 'expense', true),
('Eğlence', 'assets/icons/sinema.png', 'expense', true),
('Diğer Gider', 'assets/icons/koli.png', 'expense', true);