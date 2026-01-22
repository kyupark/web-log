package summary

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"web-log/internal/history"
)

func TagsSummary(entries []history.Entry, startDate, endDate string, days int) (string, error) {
	prompt := buildPrompt(entries, startDate, endDate, days)
	return callGemini(prompt)
}

func buildPrompt(entries []history.Entry, startDate, endDate string, days int) string {
	byDomain := map[string][]string{}
	for _, entry := range entries {
		domain := normalizeDomain(entry.URL)
		if domain == "" {
			continue
		}
		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = entry.URL
		}
		title = shortenTitle(title, 120)
		if !contains(byDomain[domain], title) {
			byDomain[domain] = append(byDomain[domain], title)
		}
	}

	domains := make([]string, 0, len(byDomain))
	for domain := range byDomain {
		domains = append(domains, domain)
	}
	sort.Strings(domains)

	lines := []string{
		fmt.Sprintf("Browsing history from %s to %s (%d days):", startDate, endDate, days),
		"",
	}
	for _, domain := range domains {
		lines = append(lines, "- "+domain)
		for _, title := range byDomain[domain] {
			lines = append(lines, "  - "+title)
		}
	}

	activityText := strings.Join(lines, "\n")

	prompt := `You are summarizing browsing history into a structured tag summary.

Rules:
- Output Markdown only.
- Start with: "# Browsing Summary - {start} to {end} ({days} days)".
- Sections: "**Shopping**" first, then "**Other**" for everything else.
- DO NOT create Social or Video sections; group by topic/tag instead.
- Group by topic/tag, NOT by site. Site is secondary info.
- Each line format: "#tag (COUNT) action text [site1.com, site2.com]".
- Tags must be in descending COUNT; #misc must be last.
- Use sites without www; for local URLs use just [localhost].
- Ignore trivial auth/redirect/login/consent pages.
- Do not include Gmail.
- #misc should be minimal; re-categorize Amazon, Digitec, finance.yahoo.com, etc. into sensible tags.
- For shopping (Amazon, Digitec, ebay, etc.), prefer phrasing like "compared Garmin watches".
- Provide action phrases + examples, not just raw titles.
  Example: "#k-content (12) watched K-content clips like Davichi vlog, Okinawa Summer Escape; … [youtube.com]"
- Keep titles concise; shorten long ones.

Now produce the summary based on the browsing history below.

` + activityText

	prompt = strings.ReplaceAll(prompt, "{start}", startDate)
	prompt = strings.ReplaceAll(prompt, "{end}", endDate)
	prompt = strings.ReplaceAll(prompt, "{days}", fmt.Sprintf("%d", days))
	return prompt
}

func normalizeDomain(url string) string {
	parts := strings.Split(url, "://")
	host := parts[len(parts)-1]
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	host = strings.ToLower(host)
	host = strings.TrimPrefix(host, "www.")
	// Strip port from localhost URLs (localhost:3000 -> localhost)
	if strings.HasPrefix(host, "localhost:") {
		host = "localhost"
	}
	return host
}

func shortenTitle(title string, maxLen int) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	for strings.Contains(title, "  ") {
		title = strings.ReplaceAll(title, "  ", " ")
	}
	title = strings.ReplaceAll(title, " - YouTube", "")
	title = strings.ReplaceAll(title, " | YouTube", "")
	title = strings.ReplaceAll(title, " / X", "")
	title = strings.ReplaceAll(title, " | X", "")
	if strings.Contains(title, " on X: ") {
		parts := strings.Split(title, " on X: ")
		title = parts[len(parts)-1]
	}
	if len(title) <= maxLen {
		return title
	}
	return strings.TrimSpace(title[:maxLen-1]) + "…"
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func DateRange(days int, from string, to string) (time.Time, time.Time, string, string, int, error) {
	now := time.Now().UTC()
	if days > 0 && (from != "" || to != "") {
		return time.Time{}, time.Time{}, "", "", 0, fmt.Errorf("--days cannot be used with --from/--to")
	}
	if to != "" && from == "" {
		return time.Time{}, time.Time{}, "", "", 0, fmt.Errorf("--to requires --from")
	}

	if days == 0 && from == "" && to == "" {
		days = 7
	}

	if days > 0 {
		endDate := startOfDay(now)
		startDate := endDate.AddDate(0, 0, -days)
		return startDate, endDate.AddDate(0, 0, 1), startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), days, nil
	}

	start, err := time.ParseInLocation("2006-01-02", from, time.UTC)
	if err != nil {
		return time.Time{}, time.Time{}, "", "", 0, fmt.Errorf("invalid --from date")
	}
	endDate := now
	if to != "" {
		endDate, err = time.ParseInLocation("2006-01-02", to, time.UTC)
		if err != nil {
			return time.Time{}, time.Time{}, "", "", 0, fmt.Errorf("invalid --to date")
		}
	}
	actualDays := int(endDate.Sub(start).Hours()/24) + 1
	return start, endDate.AddDate(0, 0, 1), start.Format("2006-01-02"), endDate.Format("2006-01-02"), actualDays, nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
