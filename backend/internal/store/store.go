package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/maquino96/timeline/internal/models"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		interval_secs INTEGER NOT NULL DEFAULT 300,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		keywords TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS items (
		id TEXT PRIMARY KEY,
		source_id INTEGER NOT NULL,
		source_type TEXT NOT NULL,
		source_name TEXT NOT NULL,
		title TEXT NOT NULL DEFAULT '',
		body TEXT NOT NULL DEFAULT '',
		url TEXT NOT NULL DEFAULT '',
		author TEXT NOT NULL DEFAULT '',
		published_at TEXT NOT NULL,
		fetched_at TEXT NOT NULL DEFAULT (datetime('now')),
		metadata TEXT NOT NULL DEFAULT '{}'
	);

	CREATE INDEX IF NOT EXISTS idx_items_published_at ON items(published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_items_source_id ON items(source_id);
	CREATE INDEX IF NOT EXISTS idx_items_source_type ON items(source_type);

	CREATE TABLE IF NOT EXISTS item_topics (
		item_id TEXT NOT NULL,
		topic_id INTEGER NOT NULL,
		PRIMARY KEY (item_id, topic_id),
		FOREIGN KEY (item_id) REFERENCES items(id),
		FOREIGN KEY (topic_id) REFERENCES topics(id)
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS items_fts USING fts5(
		title, body, content='items', content_rowid='rowid'
	);

	CREATE TRIGGER IF NOT EXISTS items_ai AFTER INSERT ON items BEGIN
		INSERT INTO items_fts(rowid, title, body) VALUES (new.rowid, new.title, new.body);
	END;

	CREATE TRIGGER IF NOT EXISTS items_ad AFTER DELETE ON items BEGIN
		INSERT INTO items_fts(items_fts, rowid, title, body) VALUES ('delete', old.rowid, old.title, old.body);
	END;
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) AddSource(source *models.Source) error {
	_, err := s.db.Exec(
		`INSERT INTO sources (type, name, url, interval_secs, enabled) VALUES (?, ?, ?, ?, ?)`,
		source.Type, source.Name, source.URL, source.Interval, source.Enabled,
	)
	return err
}

func (s *Store) GetSources() ([]models.Source, error) {
	rows, err := s.db.Query(`SELECT id, type, name, url, interval_secs, enabled, created_at FROM sources ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var src models.Source
		var createdAt string
		if err := rows.Scan(&src.ID, &src.Type, &src.Name, &src.URL, &src.Interval, &src.Enabled, &createdAt); err != nil {
			return nil, err
		}
		src.CreatedAt, _ = parseSQLiteTime(createdAt)
		sources = append(sources, src)
	}
	return sources, rows.Err()
}

func (s *Store) GetEnabledSources() ([]models.Source, error) {
	rows, err := s.db.Query(`SELECT id, type, name, url, interval_secs, enabled, created_at FROM sources WHERE enabled = 1 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var src models.Source
		var createdAt string
		if err := rows.Scan(&src.ID, &src.Type, &src.Name, &src.URL, &src.Interval, &src.Enabled, &createdAt); err != nil {
			return nil, err
		}
		src.CreatedAt, _ = parseSQLiteTime(createdAt)
		sources = append(sources, src)
	}
	return sources, rows.Err()
}

func (s *Store) UpdateSource(source *models.Source) error {
	_, err := s.db.Exec(
		`UPDATE sources SET name=?, url=?, interval_secs=?, enabled=? WHERE id=?`,
		source.Name, source.URL, source.Interval, source.Enabled, source.ID,
	)
	return err
}

func (s *Store) DeleteSource(id int64) error {
	_, err := s.db.Exec(`DELETE FROM sources WHERE id=?`, id)
	return err
}

func (s *Store) AddTopic(topic *models.Topic) error {
	_, err := s.db.Exec(
		`INSERT INTO topics (name, keywords, enabled) VALUES (?, ?, ?)`,
		topic.Name, topic.Keywords, topic.Enabled,
	)
	return err
}

func (s *Store) GetTopics() ([]models.Topic, error) {
	rows, err := s.db.Query(`SELECT id, name, keywords, enabled, created_at FROM topics ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		var createdAt string
		if err := rows.Scan(&t.ID, &t.Name, &t.Keywords, &t.Enabled, &createdAt); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = parseSQLiteTime(createdAt)
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

func (s *Store) GetEnabledTopics() ([]models.Topic, error) {
	rows, err := s.db.Query(`SELECT id, name, keywords, enabled, created_at FROM topics WHERE enabled = 1 ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		var createdAt string
		if err := rows.Scan(&t.ID, &t.Name, &t.Keywords, &t.Enabled, &createdAt); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = parseSQLiteTime(createdAt)
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

func (s *Store) UpdateTopic(topic *models.Topic) error {
	_, err := s.db.Exec(
		`UPDATE topics SET name=?, keywords=?, enabled=? WHERE id=?`,
		topic.Name, topic.Keywords, topic.Enabled, topic.ID,
	)
	return err
}

func (s *Store) DeleteTopic(id int64) error {
	_, err := s.db.Exec(`DELETE FROM topics WHERE id=?`, id)
	return err
}

func (s *Store) UpsertItem(item *models.Item) error {
	_, err := s.db.Exec(
		`INSERT INTO items (id, source_id, source_type, source_name, title, body, url, author, published_at, fetched_at, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   source_name=excluded.source_name, title=excluded.title, body=excluded.body, url=excluded.url,
		   author=excluded.author, metadata=excluded.metadata`,
		item.ID, item.SourceID, item.SourceType, item.SourceName,
		item.Title, item.Body, item.URL, item.Author,
		item.PublishedAt.Format(time.RFC3339), item.FetchedAt.Format(time.RFC3339),
		item.Metadata,
	)
	return err
}

func (s *Store) AddItemTopic(itemID string, topicID int64) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO item_topics (item_id, topic_id) VALUES (?, ?)`,
		itemID, topicID,
	)
	return err
}

func (s *Store) GetItems(limit, offset int) ([]models.Item, error) {
	rows, err := s.db.Query(
		`SELECT id, source_id, source_type, source_name, title, body, url, author, published_at, fetched_at, COALESCE(metadata, '{}')
		 FROM items ORDER BY published_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (s *Store) CountItems() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&count)
	return count, err
}

func (s *Store) GetItemsByTopic(topicID int64, limit, offset int) ([]models.Item, error) {
	rows, err := s.db.Query(
		`SELECT i.id, i.source_id, i.source_type, i.source_name, i.title, i.body, i.url, i.author, i.published_at, i.fetched_at, COALESCE(i.metadata, '{}')
		 FROM items i
		 JOIN item_topics it ON i.id = it.item_id
		 WHERE it.topic_id = ?
		 ORDER BY i.published_at DESC LIMIT ? OFFSET ?`,
		topicID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (s *Store) CountItemsByTopic(topicID int64) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM items i
		 JOIN item_topics it ON i.id = it.item_id
		 WHERE it.topic_id = ?`, topicID,
	).Scan(&count)
	return count, err
}

func (s *Store) GetItemsBySource(sourceID int64, limit, offset int) ([]models.Item, error) {
	rows, err := s.db.Query(
		`SELECT id, source_id, source_type, source_name, title, body, url, author, published_at, fetched_at, COALESCE(metadata, '{}')
		 FROM items WHERE source_id = ? ORDER BY published_at DESC LIMIT ? OFFSET ?`,
		sourceID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (s *Store) CountItemsBySource(sourceID int64) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM items WHERE source_id = ?`, sourceID).Scan(&count)
	return count, err
}

func (s *Store) GetItemsByType(sourceType string, limit, offset int) ([]models.Item, error) {
	rows, err := s.db.Query(
		`SELECT id, source_id, source_type, source_name, title, body, url, author, published_at, fetched_at, COALESCE(metadata, '{}')
		 FROM items WHERE source_type = ? ORDER BY published_at DESC LIMIT ? OFFSET ?`,
		sourceType, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (s *Store) CountItemsByType(sourceType string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM items WHERE source_type = ?`, sourceType).Scan(&count)
	return count, err
}

func (s *Store) SearchItems(query string, limit, offset int) ([]models.Item, error) {
	rows, err := s.db.Query(
		`SELECT i.id, i.source_id, i.source_type, i.source_name, i.title, i.body, i.url, i.author, i.published_at, i.fetched_at, COALESCE(i.metadata, '{}')
		 FROM items i
		 JOIN items_fts fts ON i.rowid = fts.rowid
		 WHERE items_fts MATCH ?
		 ORDER BY rank LIMIT ? OFFSET ?`,
		query, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (s *Store) CountSearchItems(query string) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM items i
		 JOIN items_fts fts ON i.rowid = fts.rowid
		 WHERE items_fts MATCH ?`, query,
	).Scan(&count)
	return count, err
}

func parseSQLiteTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

func scanItems(rows *sql.Rows) ([]models.Item, error) {
	var items []models.Item
	for rows.Next() {
		var item models.Item
		var pubStr, fetchStr string
		if err := rows.Scan(&item.ID, &item.SourceID, &item.SourceType, &item.SourceName,
			&item.Title, &item.Body, &item.URL, &item.Author,
			&pubStr, &fetchStr, &item.Metadata); err != nil {
			return nil, err
		}
		item.PublishedAt, _ = time.Parse(time.RFC3339, pubStr)
		item.FetchedAt, _ = time.Parse(time.RFC3339, fetchStr)
		items = append(items, item)
	}
	return items, rows.Err()
}
