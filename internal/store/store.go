package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "rustler.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS scans (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    total_links INTEGER DEFAULT 0,
    broken_links INTEGER DEFAULT 0,
    ssl_issues INTEGER DEFAULT 0,
    results_json TEXT DEFAULT '[]',
    started_at TEXT DEFAULT '',
    completed_at TEXT DEFAULT '',
    created_at TEXT DEFAULT (datetime('now'))
);`)
	return err
}

type Scan struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	TotalLinks  int    `json:"total_links"`
	BrokenLinks int    `json:"broken_links"`
	SSLIssues   int    `json:"ssl_issues"`
	ResultsJSON string `json:"results,omitempty"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func (db *DB) CreateScan(url string) (*Scan, error) {
	id := "scn_" + genID(8)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.conn.Exec("INSERT INTO scans (id,url,status,started_at,created_at) VALUES (?,?,'running',?,?)", id, url, now, now)
	if err != nil {
		return nil, err
	}
	return &Scan{ID: id, URL: url, Status: "running", StartedAt: now, CreatedAt: now}, nil
}

func (db *DB) CompleteScan(id string, total, broken, ssl int, resultsJSON string) {
	now := time.Now().UTC().Format(time.RFC3339)
	db.conn.Exec("UPDATE scans SET status='completed', total_links=?, broken_links=?, ssl_issues=?, results_json=?, completed_at=? WHERE id=?",
		total, broken, ssl, resultsJSON, now, id)
}

func (db *DB) FailScan(id, errMsg string) {
	now := time.Now().UTC().Format(time.RFC3339)
	db.conn.Exec("UPDATE scans SET status='failed', results_json=?, completed_at=? WHERE id=?",
		fmt.Sprintf(`[{"error":"%s"}]`, errMsg), now, id)
}

func (db *DB) GetScan(id string) (*Scan, error) {
	var s Scan
	err := db.conn.QueryRow("SELECT id,url,status,total_links,broken_links,ssl_issues,results_json,started_at,completed_at,created_at FROM scans WHERE id=?", id).
		Scan(&s.ID, &s.URL, &s.Status, &s.TotalLinks, &s.BrokenLinks, &s.SSLIssues, &s.ResultsJSON, &s.StartedAt, &s.CompletedAt, &s.CreatedAt)
	return &s, err
}

func (db *DB) ListScans() ([]Scan, error) {
	rows, err := db.conn.Query("SELECT id,url,status,total_links,broken_links,ssl_issues,started_at,completed_at,created_at FROM scans ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Scan
	for rows.Next() {
		var s Scan
		rows.Scan(&s.ID, &s.URL, &s.Status, &s.TotalLinks, &s.BrokenLinks, &s.SSLIssues, &s.StartedAt, &s.CompletedAt, &s.CreatedAt)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) DeleteScan(id string) error {
	_, err := db.conn.Exec("DELETE FROM scans WHERE id=?", id)
	return err
}

func (db *DB) Stats() map[string]any {
	var scans, broken int
	db.conn.QueryRow("SELECT COUNT(*) FROM scans").Scan(&scans)
	db.conn.QueryRow("SELECT COALESCE(SUM(broken_links),0) FROM scans").Scan(&broken)
	return map[string]any{"scans": scans, "total_broken_found": broken}
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
