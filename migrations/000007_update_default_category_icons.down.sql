-- Default kategorileri eski emoji ikonlarına geri al
UPDATE categories SET icon = '💰' WHERE is_default = true AND name = 'Maaş';
UPDATE categories SET icon = '💻' WHERE is_default = true AND name = 'Freelance';
UPDATE categories SET icon = '📈' WHERE is_default = true AND name = 'Yatırım';
UPDATE categories SET icon = '💵' WHERE is_default = true AND name = 'Diğer Gelir';
UPDATE categories SET icon = '🍔' WHERE is_default = true AND name = 'Yemek';
UPDATE categories SET icon = '🚗' WHERE is_default = true AND name = 'Ulaşım';
UPDATE categories SET icon = '🛒' WHERE is_default = true AND name = 'Alışveriş';
UPDATE categories SET icon = '📄' WHERE is_default = true AND name = 'Faturalar';
UPDATE categories SET icon = '🏥' WHERE is_default = true AND name = 'Sağlık';
UPDATE categories SET icon = '🎬' WHERE is_default = true AND name = 'Eğlence';
UPDATE categories SET icon = '📦' WHERE is_default = true AND name = 'Diğer Gider';
