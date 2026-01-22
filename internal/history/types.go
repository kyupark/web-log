package history

import "time"

type Entry struct {
	URL       string
	Title     string
	VisitTime time.Time
	Source    string
}

func Deduplicate(entries []Entry) []Entry {
	if len(entries) == 0 {
		return entries
	}
	seen := make(map[string]Entry, len(entries))
	for _, entry := range entries {
		if entry.URL == "" {
			continue
		}
		existing, ok := seen[entry.URL]
		if !ok || entry.VisitTime.After(existing.VisitTime) {
			seen[entry.URL] = entry
		}
	}
	result := make([]Entry, 0, len(seen))
	for _, entry := range seen {
		result = append(result, entry)
	}
	return result
}
