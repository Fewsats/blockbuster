CREATE TABLE invoice_status (
    payment_hash TEXT PRIMARY KEY,
    settled BOOLEAN NOT NULL,
    preimage TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_invoice_status_updated_at ON invoice_status(updated_at);