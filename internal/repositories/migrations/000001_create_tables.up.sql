BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS users
(
    id            BIGSERIAL PRIMARY KEY,
    login         VARCHAR(255) NOT NULL UNIQUE,
    password_hash BYTEA        NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders
(
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    number      TEXT           NOT NULL UNIQUE,
    status      TEXT           NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    accrual     NUMERIC(20, 2) NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id_uploaded_at ON orders (user_id, uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status_uploaded_at ON orders (status, uploaded_at);
CREATE INDEX IF NOT EXISTS idx_orders_poll ON orders (status, uploaded_at) WHERE status IN ('NEW','PROCESSING');

CREATE TABLE IF NOT EXISTS withdrawals
(
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    order_number TEXT           NOT NULL UNIQUE,
    sum          NUMERIC(20, 2) NOT NULL CHECK (sum > 0),
    processed_at TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_processed_at ON withdrawals (user_id, processed_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uidx_withdrawals_order ON withdrawals (order_number);

COMMIT;