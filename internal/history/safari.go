package history

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var safariEpoch = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

func SafariHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Safari", "History.db"), nil
}

func ReadSafariHistory(since *time.Time, until *time.Time) ([]Entry, error) {
	path, err := SafariHistoryPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp("", "weblog-safari-*.db")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer func() {
		_ = os.Remove(tmpPath)
		_ = os.Remove(tmpPath + "-wal")
		_ = os.Remove(tmpPath + "-shm")
	}()

	if err := copyFile(path, tmpPath); err != nil {
		return nil, err
	}
	_ = copyFile(path+"-wal", tmpPath+"-wal")
	_ = copyFile(path+"-shm", tmpPath+"-shm")

	db, err := sql.Open("sqlite", tmpPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	hasItemTitle, hasVisitTitle, err := safariTitleColumns(db)
	if err != nil {
		return nil, err
	}

	titleExpr := "NULL"
	switch {
	case hasItemTitle && hasVisitTitle:
		titleExpr = "COALESCE(hv.title, hi.title)"
	case hasVisitTitle:
		titleExpr = "hv.title"
	case hasItemTitle:
		titleExpr = "hi.title"
	}

	query := "SELECT hi.url, " + titleExpr + " as title, hv.visit_time FROM history_visits hv JOIN history_items hi ON hv.history_item = hi.id"
	args := []any{}
	conditions := []string{}
	if since != nil {
		sinceVal := since.Sub(safariEpoch).Seconds()
		conditions = append(conditions, "hv.visit_time >= ?")
		args = append(args, sinceVal)
	}
	if until != nil {
		untilVal := until.Sub(safariEpoch).Seconds()
		conditions = append(conditions, "hv.visit_time < ?")
		args = append(args, untilVal)
	}
	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions)
	}
	query += " ORDER BY hv.visit_time DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []Entry{}
	for rows.Next() {
		var url string
		var title sql.NullString
		var visitRaw float64
		if err := rows.Scan(&url, &title, &visitRaw); err != nil {
			return nil, err
		}
		visitTime := safariEpoch.Add(time.Duration(visitRaw * float64(time.Second)))
		entry := Entry{
			URL:       url,
			Title:     title.String,
			VisitTime: visitTime,
			Source:    "safari",
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func safariTitleColumns(db *sql.DB) (bool, bool, error) {
	hasItemTitle, err := columnExists(db, "history_items", "title")
	if err != nil {
		return false, false, err
	}
	hasVisitTitle, err := columnExists(db, "history_visits", "title")
	if err != nil {
		return false, false, err
	}
	return hasItemTitle, hasVisitTitle, nil
}

func columnExists(db *sql.DB, table string, column string) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = output.Close()
	}()

	_, err = output.ReadFrom(input)
	if err != nil {
		return err
	}
	return nil
}

func joinConditions(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	out := conditions[0]
	for i := 1; i < len(conditions); i++ {
		out += " AND " + conditions[i]
	}
	return out
}

var ErrSafariPermission = errors.New("cannot access Safari history; Full Disk Access is required")
