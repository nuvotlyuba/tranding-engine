CREATE EXTENSION IF NOT EXISTS timescaledb;
-- orders — все ордера системы
CREATE TABLE orders (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL,
    symbol      VARCHAR(20) NOT NULL,       -- "BTCUSDT"
    side        VARCHAR(4) NOT NULL,        -- "buy" / "sell"
    type        VARCHAR(10) NOT NULL,       -- "limit" / "market"
    price       NUMERIC(20, 8) NOT NULL,    -- 0 для market
    quantity    NUMERIC(20, 8) NOT NULL,
    filled      NUMERIC(20, 8) NOT NULL DEFAULT 0,
    status      VARCHAR(20) NOT NULL,       -- "open" / "filled" / "cancelled"
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- индексы для частых запросов
CREATE INDEX idx_orders_symbol_status ON orders(symbol, status);
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- trades — факты исполненных сделок
CREATE TABLE trades (
    id            UUID PRIMARY KEY,
    symbol        VARCHAR(20) NOT NULL,
    buy_order_id  UUID NOT NULL REFERENCES orders(id),
    sell_order_id UUID NOT NULL REFERENCES orders(id),
    price         NUMERIC(20, 8) NOT NULL,
    quantity      NUMERIC(20, 8) NOT NULL,
    executed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trades_symbol ON trades(symbol);
CREATE INDEX idx_trades_executed_at ON trades(executed_at DESC);

-- strategies — скрипты торговых стратегий
CREATE TABLE strategies (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL,
    name        VARCHAR(100) NOT NULL,
    symbol      VARCHAR(20) NOT NULL,
    period      VARCHAR(5) NOT NULL,       -- "1m", "5m", "1h"
    script      TEXT NOT NULL,             -- JavaScript код для goja
    is_active   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- users — трейдеры
CREATE TABLE users (
    id         UUID PRIMARY KEY,
    email      VARCHAR(255) UNIQUE NOT NULL,
    balance    NUMERIC(20, 8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- candles — TimescaleDB гипертаблица (автоматически партиционируется по времени)
CREATE TABLE candles (
    symbol     VARCHAR(20) NOT NULL,
    period     VARCHAR(5) NOT NULL,
    open_time  TIMESTAMPTZ NOT NULL,
    close_time TIMESTAMPTZ NOT NULL,
    open       NUMERIC(20, 8) NOT NULL,
    high       NUMERIC(20, 8) NOT NULL,
    low        NUMERIC(20, 8) NOT NULL,
    close      NUMERIC(20, 8) NOT NULL,
    volume     NUMERIC(20, 8) NOT NULL,
    trades     INTEGER NOT NULL,
    PRIMARY KEY (symbol, period, open_time)
);

-- превращаем в гипертаблицу TimescaleDB
SELECT create_hypertable('candles', 'open_time');

-- indicators — значения индикаторов по свечам
CREATE TABLE indicators (
    symbol     VARCHAR(20) NOT NULL,
    period     VARCHAR(5) NOT NULL,
    name       VARCHAR(10) NOT NULL,       -- "EMA", "RSI", "MACD", "BB"
    timestamp  TIMESTAMPTZ NOT NULL,
    data       JSONB NOT NULL,             -- {"ema": "102.5"} или {"upper": "110", "lower": "90"}
    PRIMARY KEY (symbol, period, name, timestamp)
);

SELECT create_hypertable('indicators', 'timestamp');