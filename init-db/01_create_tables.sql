-- Создаем таблицу кошельков
CREATE TABLE IF NOT EXISTS wallets (
    id BIGSERIAL PRIMARY KEY,
    wallet_id UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- Индекс по wallet_id
CREATE INDEX IF NOT EXISTS idx_wallets_wallet_id ON wallets(wallet_id);

-- Тестовый кошелёк
INSERT INTO wallets (wallet_id, balance)
VALUES ('11111111-1111-1111-1111-111111111111', 0)
ON CONFLICT (wallet_id) DO NOTHING;