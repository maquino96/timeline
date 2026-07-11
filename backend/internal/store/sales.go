package store

import (
	"database/sql"
	"strconv"

	"github.com/maquino96/timeline/internal/models"
)

func (s *Store) GetWatchItems(activeOnly bool) ([]models.WatchItem, error) {
	query := `SELECT id, name, search_term, threshold, floor, category, active,
		ebay_price, slickdeals_price, reddit_price, last_checked, created_at, updated_at
		FROM watch_items`
	if activeOnly {
		query += " WHERE active = 1"
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWatchItems(rows)
}

func (s *Store) GetWatchItem(id int64) (*models.WatchItem, error) {
	rows, err := s.db.Query(
		`SELECT id, name, search_term, threshold, floor, category, active,
			ebay_price, slickdeals_price, reddit_price, last_checked, created_at, updated_at
			FROM watch_items WHERE id = ?`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanWatchItems(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (s *Store) AddWatchItem(item *models.WatchItem) error {
	res, err := s.db.Exec(
		`INSERT INTO watch_items (name, search_term, threshold, floor, category, active)
		 VALUES (?, ?, ?, ?, ?, 1)`,
		item.Name, item.SearchTerm, item.Threshold, item.Floor, item.Category,
	)
	if err != nil {
		return err
	}
	item.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) UpdateWatchItem(item *models.WatchItem) error {
	_, err := s.db.Exec(
		`UPDATE watch_items SET name=?, search_term=?, threshold=?, floor=?, category=?, active=?,
		 updated_at=datetime('now') WHERE id=?`,
		item.Name, item.SearchTerm, item.Threshold, item.Floor, item.Category, item.Active, item.ID,
	)
	return err
}

func (s *Store) DeleteWatchItem(id int64) error {
	_, err := s.db.Exec(`DELETE FROM sale_alerts WHERE item_id=?`, id)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`DELETE FROM watch_items WHERE id=?`, id)
	return err
}

func (s *Store) UpdateWatchItemPrice(id int64, source string, price float64) error {
	col := map[string]string{
		"ebay":       "ebay_price",
		"slickdeals": "slickdeals_price",
		"reddit":     "reddit_price",
	}[source]
	if col == "" {
		return nil
	}
	_, err := s.db.Exec(
		`UPDATE watch_items SET `+col+`=?, last_checked=datetime('now'), updated_at=datetime('now') WHERE id=?`,
		price, id,
	)
	return err
}

func (s *Store) AddSaleAlert(alert *models.SaleAlert) error {
	res, err := s.db.Exec(
		`INSERT INTO sale_alerts (item_id, price, title, deal_url, source, sent)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		alert.ItemID, alert.Price, alert.Title, alert.DealURL, alert.Source, alert.Sent,
	)
	if err != nil {
		return err
	}
	alert.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) HasRecentSaleAlert(itemID int64, hours int) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM sale_alerts WHERE item_id = ? AND created_at >= datetime('now', ?)`,
		itemID, sqlHoursOffset(hours),
	).Scan(&count)
	return count > 0, err
}

func (s *Store) DismissSaleAlert(id int64) error {
	_, err := s.db.Exec(`UPDATE sale_alerts SET dismissed = 1 WHERE id = ?`, id)
	return err
}

func (s *Store) GetRecentSaleAlerts(days, limit, offset int) ([]models.SaleAlert, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.item_id, i.name, a.price, a.title, a.deal_url, a.source, a.sent, a.dismissed, a.created_at
		 FROM sale_alerts a
		 JOIN watch_items i ON i.id = a.item_id
		 WHERE a.created_at >= datetime('now', ?) AND a.dismissed = 0
		 ORDER BY a.created_at DESC LIMIT ? OFFSET ?`,
		sqlDaysOffset(days), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.SaleAlert
	for rows.Next() {
		var a models.SaleAlert
		var createdAt string
		if err := rows.Scan(&a.ID, &a.ItemID, &a.ItemName, &a.Price, &a.Title,
			&a.DealURL, &a.Source, &a.Sent, &a.Dismissed, &createdAt); err != nil {
			return nil, err
		}
		a.CreatedAt, _ = parseSQLiteTime(createdAt)
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

func (s *Store) CountRecentSaleAlerts(days int) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM sale_alerts WHERE created_at >= datetime('now', ?) AND dismissed = 0`,
		sqlDaysOffset(days),
	).Scan(&count)
	return count, err
}

func sqlDaysOffset(days int) string {
	return "-" + strconv.Itoa(days) + " days"
}

func sqlHoursOffset(hours int) string {
	return "-" + strconv.Itoa(hours) + " hours"
}

func scanWatchItems(rows *sql.Rows) ([]models.WatchItem, error) {
	var items []models.WatchItem
	for rows.Next() {
		var it models.WatchItem
		var createdAt, updatedAt string
		if err := rows.Scan(&it.ID, &it.Name, &it.SearchTerm, &it.Threshold, &it.Floor,
			&it.Category, &it.Active, &it.EbayPrice, &it.SlickdealsPrice, &it.RedditPrice,
			&it.LastChecked, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		it.CreatedAt, _ = parseSQLiteTime(createdAt)
		it.UpdatedAt, _ = parseSQLiteTime(updatedAt)
		items = append(items, it)
	}
	return items, rows.Err()
}
