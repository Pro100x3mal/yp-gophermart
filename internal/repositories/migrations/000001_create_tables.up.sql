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
    user_id     BIGINT     NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    number      TEXT        NOT NULL UNIQUE,
    status      TEXT        NOT NULL DEFAULT 'NEW' CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    accrual     BIGINT NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id_uploaded_at ON orders (user_id, uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status_uploaded_at ON orders (status, uploaded_at);

COMMIT;