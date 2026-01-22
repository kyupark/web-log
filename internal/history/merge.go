package history

import "time"

func ReadAllHistory(since *time.Time, until *time.Time) ([]Entry, []error) {
	entries := []Entry{}
	errors := []error{}

	safariEntries, err := ReadSafariHistory(since, until)
	if err != nil {
		errors = append(errors, err)
	} else {
		entries = append(entries, safariEntries...)
	}

	chromeEntries, err := ReadChromeHistory(since, until)
	if err != nil {
		errors = append(errors, err)
	} else {
		entries = append(entries, chromeEntries...)
	}

	return entries, errors
}
