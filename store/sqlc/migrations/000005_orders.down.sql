DELETE INDEX IF EXISTS purchases_external_id_idx;
DELETE INDEX IF EXISTS purchases_user_id_idx;
DROP TABLE IF EXISTS purchases;

DELETE INDEX IF EXISTS offers_external_id_idx;
DELETE INDEX IF EXISTS offers_payment_hash_idx;
DELETE INDEX IF EXISTS offers_created_at_idx;
DROP TABLE IF EXISTS offers;
