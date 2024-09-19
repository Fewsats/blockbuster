CREATE TABLE videos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id TEXT NOT NULL,
    user_email TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    video_url TEXT NOT NULL,
    cover_url TEXT NOT NULL,
    price_in_cents INTEGER NOT NULL,
    total_views INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    -- TODO(pol) FK is not enabled as we support direct video upload
    -- we might want to auto-create the user before creating the video & enforce FK
    -- FOREIGN KEY (user_email) REFERENCES users(email)
);