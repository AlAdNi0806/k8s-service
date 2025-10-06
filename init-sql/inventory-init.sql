CREATE TABLE IF NOT EXISTS stock (
    product_id BIGINT PRIMARY KEY,
    quantity INT NOT NULL CHECK (quantity >= 0)
);

-- Предзаполнение для демо (опционально)
INSERT INTO stock (product_id, quantity)
VALUES (123, 100)
ON CONFLICT (product_id) DO NOTHING;
