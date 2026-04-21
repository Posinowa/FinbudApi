-- Mevcut default kategorilerin emoji ikonlarını asset path'lerine güncelle
UPDATE categories SET icon = 'assets/icons/para.png'           WHERE is_default = true AND name = 'Maaş';
UPDATE categories SET icon = 'assets/icons/laptop.png'         WHERE is_default = true AND name = 'Freelance';
UPDATE categories SET icon = 'assets/icons/para_akis.png'      WHERE is_default = true AND name = 'Yatırım';
UPDATE categories SET icon = 'assets/icons/gelir_cüzdan.png'   WHERE is_default = true AND name = 'Diğer Gelir';
UPDATE categories SET icon = 'assets/icons/restorant.png'      WHERE is_default = true AND name = 'Yemek';
UPDATE categories SET icon = 'assets/icons/araba.png'          WHERE is_default = true AND name = 'Ulaşım';
UPDATE categories SET icon = 'assets/icons/market_arabasi.png' WHERE is_default = true AND name = 'Alışveriş';
UPDATE categories SET icon = 'assets/icons/fatura.png'         WHERE is_default = true AND name = 'Faturalar';
UPDATE categories SET icon = 'assets/icons/saglik.png'         WHERE is_default = true AND name = 'Sağlık';
UPDATE categories SET icon = 'assets/icons/sinema.png'         WHERE is_default = true AND name = 'Eğlence';
UPDATE categories SET icon = 'assets/icons/koli.png'           WHERE is_default = true AND name = 'Diğer Gider';
