-- offers is a table that stores the offers for the services.
CREATE TABLE IF NOT EXISTS offers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- user_id is the user ID of the user selling this service.
    user_id BIGINT NOT NULL REFERENCES users(id),

    -- external_id is the external ID of the service.
    external_id TEXT NOT NULL,

    -- payment_hash is the payment hash for the offer.
    payment_hash TEXT UNIQUE NOT NULL,

    -- price_in_cents is the price of the item
    price_in_cents BIGINT NOT NULL,

    -- currency is the currency of the item.
    currency TEXT NOT NULL,

    -- expiration_date is the expiration date for the credentials linked to this
    -- offer.
    expiration_date DATETIME,

    -- created_at is the timestamp when the offer was created.
    created_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS offers_external_id_idx ON offers (external_id);
CREATE INDEX IF NOT EXISTS offers_payment_hash_idx ON offers (payment_hash);
CREATE INDEX IF NOT EXISTS offers_created_at_idx ON offers (created_at);


-- purchases is a table that stores the purchases made by the end clients.
CREATE TABLE IF NOT EXISTS purchases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- user_id is the user ID of the user selling this service.
    user_id BIGINT NOT NULL REFERENCES users(id),

    -- external_id is the external ID of the service.
    external_id TEXT NOT NULL,

    -- service_type is the type of the service. (STaaS, DaaS, FaaS...)
    service_type TEXT NOT NULL,

    -- price_in_cents is the price of the item in cents.
    price_in_cents BIGINT NOT NULL,

    -- currency is the currency used for the transaction.
    currency TEXT NOT NULL,

    -- expiration_date is the expiration date for the credentials linked to this
    -- purchase.
    expiration_date DATETIME,

    -- payment_hash is the payment hash for the offer linked to this
    -- purchase.
    payment_hash TEXT NOT NULL UNIQUE REFERENCES offers(payment_hash),

    -- created_at is the timestamp when the purchase was created.
    created_at DATETIME NOT NULL

);

CREATE INDEX IF NOT EXISTS purchases_user_id_idx ON purchases (user_id);
CREATE INDEX IF NOT EXISTS purchases_external_id_idx ON purchases (external_id);
