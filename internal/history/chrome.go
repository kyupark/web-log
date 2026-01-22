package history

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var chromeEpoch = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

func ChromeHistoryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "History"), nil
}

func ReadChromeHistory(since *time.Time, until *time.Time) ([]Entry, error) {
	path, err := ChromeHistoryPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp("", "weblog-chrome-*.db")
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

	query := "SELECT urls.url, urls.title, visits.visit_time FROM visits JOIN urls ON visits.url = urls.id"
	args := []any{}
	conditions := []string{}
	if since != nil {
		sinceVal := since.Sub(chromeEpoch).Microseconds()
		conditions = append(conditions, "visits.visit_time >= ?")
		args = append(args, sinceVal)
	}
	if until != nil {
		untilVal := until.Sub(chromeEpoch).Microseconds()
		conditions = append(conditions, "visits.visit_time < ?")
		args = append(args, untilVal)
	}
	if len(conditions) > 0 {
		query += " WHERE " + joinConditions(conditions)
	}
	query += " ORDER BY visits.visit_time DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []Entry{}
	for rows.Next() {
		var url string
		var title sql.NullString
		var visitRaw int64
		if err := rows.Scan(&url, &title, &visitRaw); err != nil {
			return nil, err
		}
		visitTime := chromeEpoch.Add(time.Duration(visitRaw) * time.Microsecond)
		entries = append(entries, Entry{
			URL:       url,
			Title:     title.String,
			VisitTime: visitTime,
			Source:    "chrome",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
