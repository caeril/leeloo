package shell

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/caeril/leeloo/data"
)

func _getClockedEntries(tomb *data.Tomb) []data.Entry {
	m := make(map[string]data.Entry)
	for _, entry := range tomb.Entries {
		existingEntry, haz := m[entry.Key]
		if !haz {
			m[entry.Key] = entry
		} else if entry.Clock > existingEntry.Clock {
			m[entry.Key] = entry
		}
	}
	out := []data.Entry{}
	for _, entry := range m {
		out = append(out, entry)
	}
	// todo: clean up memory here
	return out
}

func _checkAllVersions(toks []string) (bool, []string) {
	allVersions := false
	out := []string{}
	for _, tok := range toks {
		if tok == "--all" {
			allVersions = true
		} else {
			out = append(out, tok)
		}
	}
	return allVersions, out
}

func _checkShowDeleted(toks []string) (bool, []string) {
	showDeleted := false
	out := []string{}
	for _, tok := range toks {
		if tok == "--deleted" {
			showDeleted = true
		} else {
			out = append(out, tok)
		}
	}
	return showDeleted, out
}

func checkShow(command string, toks []string, tomb *data.Tomb) bool {

	if command == "list" {
		var allVersions bool
		var showDeleted bool
		var entries []data.Entry
		allVersions, toks = _checkAllVersions(toks)
		showDeleted, toks = _checkShowDeleted(toks)
		query := ""
		if len(toks) > 1 {
			query = strings.ToLower(toks[1])
		}

		if allVersions {
			entries = tomb.Entries
		} else {
			entries = _getClockedEntries(tomb)
		}

		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
		for _, entry := range entries {
			if query == "" || strings.Contains(strings.ToLower(entry.Key), query) {
				flags := ""
				if 0 != (entry.Flags & data.EntryFlagDeleted) {
					flags += "D"
					if !showDeleted {
						continue
					}
				}
				fmt.Fprintf(w, " ^ %s\t v %d (%s)\n", entry.Key, entry.Clock, flags)
			}
		}
		w.Flush()
		fmt.Println()
		return true

	}

	if command == "show" {
		var allVersions bool
		var showDeleted bool
		var entries []data.Entry
		allVersions, toks = _checkAllVersions(toks)
		showDeleted, toks = _checkShowDeleted(toks)
		query := ""
		if len(toks) > 1 {
			query = strings.ToLower(toks[1])
		}

		if allVersions {
			entries = tomb.Entries
		} else {
			entries = _getClockedEntries(tomb)
		}

		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
		for _, entry := range entries {
			if query == "" || strings.Contains(strings.ToLower(entry.Key), query) {
				flags := ""
				if 0 != (entry.Flags & data.EntryFlagDeleted) {
					flags += "D"
					if !showDeleted {
						continue
					}
				}
				fmt.Fprintf(w, " ^ %s\t %s\t v %d (%s)\n", entry.Key, entry.Value, entry.Clock, flags)
			}
		}
		w.Flush()
		fmt.Println()
		return true
	}
	// todo: clean up memory here

	return false
}
