package summary

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"web-log/internal/history"
)

func TagsSummary(entries []history.Entry, startDate, endDate string, days int) (string, error) {
	// Filter out noise (domain-level only)
	filtered := make([]history.Entry, 0, len(entries))
	for _, entry := range entries {
		domain := normalizeDomain(entry.URL)
		// Skip by domain
		if domain == "mail.google.com" ||
			domain == "accounts.google.com" ||
			strings.HasPrefix(domain, "auth.") ||
			strings.HasPrefix(domain, "login.") ||
			strings.HasPrefix(domain, "sso.") {
			continue
		}
		filtered = append(filtered, entry)
	}

	prompt := buildPrompt(filtered, startDate, endDate, days)
	return callOpenRouter(prompt)
}

func buildPrompt(entries []history.Entry, startDate, endDate string, days int) string {
	// Group entries by date
	byDate := map[string][]history.Entry{}
	for _, entry := range entries {
		date := entry.VisitTime.Format("2006-01-02")
		byDate[date] = append(byDate[date], entry)
	}

	// Sort dates
	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	var lines []string
	lines = append(lines, fmt.Sprintf("Browsing history from %s to %s (%d days):", startDate, endDate, days))
	lines = append(lines, "")

	for _, date := range dates {
		lines = append(lines, "## "+date)
		lines = append(lines, "time | url | title")
		lines = append(lines, "--- | --- | ---")

		// Sort entries by time within the day
		dayEntries := byDate[date]
		sort.Slice(dayEntries, func(i, j int) bool {
			return dayEntries[i].VisitTime.Before(dayEntries[j].VisitTime)
		})

		for _, entry := range dayEntries {
			timeStr := entry.VisitTime.Format("15:04")
			url := entry.URL
			if len(url) > 100 {
				url = url[:97] + "..."
			}
			title := strings.TrimSpace(entry.Title)
			if title == "" {
				title = "-"
			}
			title = shortenTitle(title, 80)
			// Escape pipe characters in URL and title
			url = strings.ReplaceAll(url, "|", "%7C")
			title = strings.ReplaceAll(title, "|", "-")
			lines = append(lines, fmt.Sprintf("%s | %s | %s", timeStr, url, title))
		}
		lines = append(lines, "")
	}

	activityText := strings.Join(lines, "\n")

	prompt := `You are summarizing browsing history into a structured tag summary for personal journaling.

Rules:
- Output plain Markdown only. No HTML entities (no &nbsp;, &amp;, etc). Use spaces for indentation.
- Start with: "# Browsing Summary - {start} to {end} ({days} days)".
- Dynamically create sections based on topic categories found (e.g., **Shopping**, **Development**, **Research**, **Finance**).
- STRICT: Only create a section if it has 5+ items total. Merge smaller groups into the most relevant larger section.
- Group by topic/tag, NOT by site. Site is secondary info.
- Each line format: "#tag (COUNT) action text [site1.com, site2.com]".
- ABSOLUTE RULE - SUB-TAGS (STRICTLY ENFORCED):
  COUNT THE NUMBER. If the tag count is less than 10, it MUST be a single line with NO indented sub-bullets beneath it.

  ✓ CORRECT:
  #ai-agents (18) explored AI agents
    - #clawdbot (8) personal AI assistant [clawd.bot]
    - #ralph (5) autonomous coding loop [github.com/repo]
  #vscode (3) downloaded VS Code [code.visualstudio.com]
  #stocks (5) checked NVDA, TSLA charts [finance.yahoo.com]
  #books (4) read 'Atomic Habits', 'Deep Work' [ridibooks.com]

  ✗ WRONG - these have sub-bullets but count is under 10:
  #stocks (5) stock charts
    - #nvda (2) NVIDIA
  #books (4) browsed books
    - #fiction (2) fiction

- Sub-tags must sum to LESS than parent total. Skip minor items (1-2 counts).
- If sub-tags share the same site as parent, put site URL in parent only. Example:
  #youtube (29) watched various videos [m.youtube.com]
    - #korean-vlogs (10) watched vlogs about daily life
    - #running (6) watched videos about Garmin watches for runners
- Combine very small unrelated tags (1-2 counts each) into one summary line at end of section. Example:
  #misc-dev (4) downloaded VS Code, checked Remotion, explored Raycast [code.visualstudio.com, remotion.dev, raycast.com]
- Do NOT list individual webpage titles. Always group into meaningful tags.
- Tags must be in descending COUNT within each section.

Site references (IMPORTANT):
- For github.com: ALWAYS include repo path like github.com/steipete/bird, github.com/michaelshimeles/ralphy. NEVER just "github.com"
- For x.com: ALWAYS include username like x.com/steipete, x.com/clawdbot. NEVER just "x.com"
- For google maps: mention searched locations/keywords
- For local URLs: just use [localhost], no IP addresses
- NEVER repeat the same site/path in brackets.

Ignore completely (not useful for journaling):
- Authentication, account management, login pages (accounts.google.com, myaccount.google.com, sso.*, login.*, etc.)
- Redirect, consent, cookie pages
- Gmail and mail.google.com
- Unsubscribe pages, email management links (swanbitcoin unsubscribe, etc.)
- Any URL containing "unsubscribe", "email-preferences", "manage-subscription"

Categorization:
- Product research (comparing watches, reading reviews) belongs in Shopping, not Research.
- Use specific meaningful tags: #grocery for food items (not #products), #books for reading, etc.
- Avoid generic tags like #products, #misc, #general.
- IMPORTANT: Group by TOPIC, not by site. If browsing Garmin watches on ricardo.ch, include ricardo.ch in #garmin sites list - do NOT create separate #ricardo tag. The tag should reflect WHAT was browsed, not WHERE.

Content detail (CRITICAL - this is for personal journaling to remember what was done):
- Do NOT give generic descriptions. Provide SPECIFIC details that help recall what was actually consumed.
- For discussions/tweets: mention specific topics, people, or key points discussed (e.g., "Saylor acquired 22,305 BTC", "debate about risk-on vs risk-off")
- For books: list book/author names
- For videos: list specific video topics or titles
- For articles: mention specific subjects covered
- For products: list specific models compared
- The goal is to easily recall what was actually viewed/done - generic summaries are useless.
  Good: "#crypto (10) Saylor acquired 22,305 BTC at $95k, debate on BTC as risk-on vs risk-off asset, silver decade-long base breakout [x.com/saylor, x.com/JoeConsorti]"
  Bad: "#crypto (10) followed Bitcoin price discussions and metrics [x.com]"
  Good: "#books (7) read 'Atomic Habits', browsed 'Deep Work' [ridibooks.com]"
  Bad: "#ridibooks (7) ebook subscription service and books [ridibooks.com]"

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
	if strings.HasPrefix(host, "localhost:") || strings.HasPrefix(host, "127.0.0.1") || strings.HasPrefix(host, "0.0.0.0") {
		host = "localhost"
	}
	// Local network IPs (e.g., 100.x.x.x, 192.168.x.x, 10.x.x.x) -> localhost
	if len(host) > 0 && host[0] >= '0' && host[0] <= '9' {
		host = "localhost"
	}
	return host
}

func shortenTitle(title string, maxLen int) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	// Decode HTML entities (e.g., &eacute; -> é)
	title = html.UnescapeString(title)
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
