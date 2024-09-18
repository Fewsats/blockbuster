package database

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return &SQLiteDB{db: db}, nil
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

func (s *SQLiteDB) StoreToken(email, token string, expiration time.Time) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		_, err = tx.Exec("INSERT INTO users (email) VALUES (?)", email)
		if err != nil {
			return err
		}
	}

	// Store the token
	_, err = tx.Exec("INSERT INTO tokens (email, token, expiration) VALUES (?, ?, ?)", email, token, expiration)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLiteDB) VerifyToken(token string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var email string
	err = tx.QueryRow("SELECT email FROM tokens WHERE token = ? AND expiration > ?", token, time.Now()).Scan(&email)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec("DELETE FROM tokens WHERE token = ?", token)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec("UPDATE users SET verified = ? WHERE email = ?", true, email)
	if err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return email, nil
}

func (s *SQLiteDB) GetUserByEmail(email string) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	return exists, err
}

func (s *SQLiteDB) CreateUser(email string) error {
	_, err := s.db.Exec("INSERT INTO users (email) VALUES (?)", email)
	return err
}

func createTables(db *sql.DB) error {
	if err := createUsersTable(db); err != nil {
		return err
	}
	if err := createTokensTable(db); err != nil {
		return err
	}
	if err := createVideosTable(db); err != nil {
		return err
	}
	return nil
}

func createUsersTable(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT UNIQUE NOT NULL,
            verified BOOLEAN NOT NULL DEFAULT FALSE
        );
    `)
	return err
}

func createTokensTable(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS tokens (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT NOT NULL,
            token TEXT UNIQUE NOT NULL,
            expiration DATETIME NOT NULL
        );
    `)
	return err
}

func createVideosTable(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS videos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_email TEXT NOT NULL,
            title TEXT NOT NULL,
            description TEXT,
            file_path TEXT NOT NULL,
            thumbnail_path TEXT,
            price REAL NOT NULL,
            total_views INTEGER NOT NULL DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_email) REFERENCES users(email)
        );
    `)
	return err
}

func (s *SQLiteDB) CreateVideo(userEmail, title, description, filePath, thumbnailPath string, price float64) (int64, error) {
	result, err := s.db.Exec(`
        INSERT INTO videos (user_email, title, description, file_path, thumbnail_path, price)
        VALUES (?, ?, ?, ?, ?, ?)
    `, userEmail, title, description, filePath, thumbnailPath, price)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *SQLiteDB) GetVideo(id int64) (*Video, error) {
	var video Video
	err := s.db.QueryRow(`
        SELECT id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at
        FROM videos WHERE id = ?
    `, id).Scan(&video.ID, &video.UserEmail, &video.Title, &video.Description, &video.FilePath, &video.ThumbnailPath, &video.Price, &video.TotalViews, &video.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

type Video struct {
	ID            int64
	UserEmail     string
	Title         string
	Description   string
	FilePath      string
	ThumbnailPath string
	Price         float64
	TotalViews    int64
	CreatedAt     time.Time
}

func (s *SQLiteDB) GetUserVideos(userEmail string) ([]*Video, error) {
	rows, err := s.db.Query(`
        SELECT id, user_email, title, description, file_path, thumbnail_path, price, total_views, created_at
        FROM videos ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		var video Video
		err := rows.Scan(&video.ID, &video.UserEmail, &video.Title, &video.Description, &video.FilePath, &video.ThumbnailPath, &video.Price, &video.TotalViews, &video.CreatedAt)
		if err != nil {
			return nil, err
		}
		videos = append(videos, &video)
	}
	return videos, nil
}
