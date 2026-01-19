-- 1. Включаем расширение для генерации UUID (если еще нет)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. Таблица USERS
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE, -- Важно: UNIQUE!
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Таблица ACCOUNTS
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance NUMERIC(20, 4) NOT NULL DEFAULT 0, -- DECIMAL для денег!
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Один юзер может иметь только один счет в одной валюте (опционально)
    UNIQUE(user_id, currency) 
);

-- 4. Таблица TRANSACTIONS
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    to_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    amount NUMERIC(20, 4) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, success, failed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_transactions_from ON transactions(from_account_id);
CREATE INDEX idx_transactions_to ON transactions(to_account_id);

