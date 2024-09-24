-- macaroon_credentials is a table that stores the macaroon token id and root keys.
CREATE TABLE IF NOT EXISTS macaroon_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- The token_id is the unique identifier for a macaroon.
    token_id BLOB NOT NULL,

    -- The root_key is the key used in the macaroon linked to the token_id.
    root_key BLOB NOT NULL,

     -- created_at is the timestamp when the API key was created.
    created_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS macaroon_credentials_token_id ON macaroon_credentials (token_id);
