package shell

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caeril/leeloo/data"
)

func _getPriorClock(tomb *data.Tomb, key string) uint32 {
	clock := uint32(0)
	for i := range tomb.Entries {
		if tomb.Entries[i].Key == key && tomb.Entries[i].Clock > clock {
			clock = tomb.Entries[i].Clock
		}
	}
	return clock
}

func checkAdd(command string, toks []string, tomb *data.Tomb) bool {

	if command == "add" {

		if len(toks) < 3 {
			fmt.Printf(" ^ nope. need a key and a value\n")
			return true
		}

		key := toks[1]
		val := toks[2]

		if strings.HasPrefix(key, "\"") {
			fmt.Printf(" ^ i don't support quoted keys yet\n")
			return true
		}

		if strings.HasPrefix(val, "\"") {
			fmt.Printf(" ^ i don't support quoted values yet\n")
			return true
		}

		clock := _getPriorClock(tomb, key) + 1
		tomb.Entries = append(tomb.Entries, data.Entry{Clock: clock, Key: key, Value: val})
		data.SealTomb(tomb)
		fmt.Printf(" ^ key has been committed (version %d)\n", tomb.Version)
		return true

	}

	if command == "gen" {

		if len(toks) < 3 {
			fmt.Printf(" ^ nope. need a key and a length\n")
			return true
		}

		tlen, err := strconv.Atoi(toks[2])
		if err != nil {
			fmt.Printf(" ^ error: %s\n", err)
			return true
		}
		if tlen < 4 || tlen > 32 {
			fmt.Printf(" ^ invalid password length %d\n", tlen)
			return true
		}

		key := toks[1]
		val := data.CreateCSPassword(tlen)

		if strings.HasPrefix(key, "\"") {
			fmt.Printf(" ^ i don't support quoted keys yet\n")
			return true
		}

		clock := _getPriorClock(tomb, key) + 1
		tomb.Entries = append(tomb.Entries, data.Entry{Clock: clock, Key: key, Value: val})
		data.SealTomb(tomb)
		fmt.Printf(" ^ key %s has been committed with password %s (version %d)\n", key, val, tomb.Version)
		return true

	}

	return false
}
