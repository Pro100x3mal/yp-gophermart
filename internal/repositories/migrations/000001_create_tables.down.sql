BEGIN TRANSACTION;

DROP INDEX IF EXISTS uidx_withdrawals_order;
DROP INDEX IF EXISTS idx_withdrawals_user_processed_at;

DROP TABLE IF EXISTS withdrawals;

DROP INDEX IF EXISTS idx_orders_status_uploaded_at;
DROP INDEX IF EXISTS idx_orders_user_id_uploaded_at;

DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS users;

COMMIT;