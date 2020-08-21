package shell

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caeril/leeloo/data"
)

const _inactivityTimeout = int64(60) // one minute inactivity threshold

var lastActionTime int64

func runInactivityWorker() {
	for true {
		time.Sleep(time.Second * 5)
		now := time.Now().Unix()
		if now-lastActionTime > _inactivityTimeout {
			fmt.Printf("\n ^ you are too inactive. aborting.\n")
			data.Cleanup()
			os.Exit(0)
		}
	}
}

func init() {
	lastActionTime = time.Now().Unix()
	go runInactivityWorker()
}

func markAction() {
	lastActionTime = time.Now().Unix()
}

func ProcessTokens(tomb *data.Tomb, toks []string) bool {

	markAction()

	command := strings.ToLower(toks[0])

	if command == "exit" || command == "quit" {
		data.Cleanup()
		fmt.Printf(" ^ goodbye.\n")
		os.Exit(0)
	}

	if checkHelp(command, toks) {
		return true
	}

	if checkShow(command, toks, tomb) {
		return true
	}

	if checkAdd(command, toks, tomb) {
		return true
	}

	if checkSync(command, toks, tomb) {
		return true
	}

	if checkPassphrase(command, toks, tomb) {
		return true
	}

	if command == "del" {

		if len(toks) < 2 {
			fmt.Printf(" ^ nope. need a key to delete\n")
			return true
		}

		key := toks[1]
		index := -1
		for i, entry := range tomb.Entries {
			if entry.Key == key {
				index = i
				tomb.Entries[i].Flags |= data.EntryFlagDeleted
			}
		}
		if index == -1 {
			fmt.Printf(" ^ could not find key\n")
			return true
		}

		data.SealTomb(tomb)
		fmt.Printf(" ^ key deletion has been committed (version %d)\n", tomb.Version)

		return true

	}

	return false
}
