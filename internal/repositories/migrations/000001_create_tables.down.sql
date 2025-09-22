BEGIN TRANSACTION;

DROP INDEX IF EXISTS idx_orders_status_uploaded_at;
DROP INDEX IF EXISTS idx_orders_user_id_uploaded_at;

DROP TABLE IF EXISTS orders;

DROP TABLE IF EXISTS users;

COMMIT;