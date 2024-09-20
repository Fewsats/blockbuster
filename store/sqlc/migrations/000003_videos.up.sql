CREATE TABLE IF NOT EXISTS videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT NOT NULL,
    user_id INTEGER NOT NULL,

    -- l402 product info
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    cover_url TEXT NOT NULL, -- our user-uploaded over image
    price_in_cents INTEGER NOT NULL,

    -- total views represents the number of times the video has been view
    total_views INTEGER NOT NULL DEFAULT 0,
    
    -- data retrieved from cloudflare stream after upload
    thumbnail_url TEXT, -- cloudflare-generated thumbnail
    duration_in_seconds FLOAT,
    size_in_bytes INTEGER,
    input_height INTEGER,
    input_width INTEGER,
    ready_to_stream BOOLEAN NOT NULL DEFAULT FALSE,
    
    created_at DATETIME NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id)
);