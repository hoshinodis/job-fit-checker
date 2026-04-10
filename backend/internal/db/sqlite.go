package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	d, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if err := migrate(d); err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func migrate(d *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS profiles (
			id TEXT PRIMARY KEY,
			preferred_languages_json TEXT NOT NULL DEFAULT '[]',
			avoid_languages_json TEXT NOT NULL DEFAULT '[]',
			interests_json TEXT NOT NULL DEFAULT '[]',
			low_interests_json TEXT NOT NULL DEFAULT '[]',
			work_style_json TEXT NOT NULL DEFAULT '[]',
			desired_compensation TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS match_jobs (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			job_input_type TEXT NOT NULL,
			job_input_value TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'queued',
			error_message TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS extracted_job_texts (
			id TEXT PRIMARY KEY,
			match_job_id TEXT NOT NULL,
			source_url TEXT NOT NULL,
			page_title TEXT NOT NULL DEFAULT '',
			meta_description TEXT NOT NULL DEFAULT '',
			extracted_text TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (match_job_id) REFERENCES match_jobs(id)
		)`,
		`CREATE TABLE IF NOT EXISTS match_results (
			id TEXT PRIMARY KEY,
			match_job_id TEXT NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			summary TEXT NOT NULL DEFAULT '',
			pros_json TEXT NOT NULL DEFAULT '[]',
			cons_json TEXT NOT NULL DEFAULT '[]',
			questions_to_ask_json TEXT NOT NULL DEFAULT '[]',
			clipboard_text TEXT NOT NULL DEFAULT '',
			model_name TEXT NOT NULL DEFAULT '',
			raw_response TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (match_job_id) REFERENCES match_jobs(id)
		)`,
	}
	for _, s := range stmts {
		if _, err := d.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
