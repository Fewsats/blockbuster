-- macaroon_credentials is a table that stores the macaroon token id and root keys.
CREATE TABLE IF NOT EXISTS macaroon_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- The identifier is the unique identifier for a macaroon and contains the user pubkey.
    identifier TEXT NOT NULL,

    -- The root_key is the key used in the macaroon linked to the identifier.
    root_key TEXT NOT NULL,

     -- created_at is the timestamp when the API key was created.
    created_at TIMESTAMPTZ NOT NULL,

    -- encoded_base_macaroon is the base64 encoded macaroon.
    encoded_base_macaroon TEXT NOT NULL,

    -- disabled is a flag to disable the macaroon.
    disabled BOOLEAN NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS macaroon_credentials_identifier ON macaroon_credentials (identifier);
