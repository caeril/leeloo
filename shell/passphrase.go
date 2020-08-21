package shell

import (
	"fmt"
	"syscall"

	"github.com/caeril/leeloo/data"
	"github.com/caeril/leeloo/paranoia"
	"golang.org/x/crypto/ssh/terminal"
)

func checkPassphrase(command string, tokens []string, tomb *data.Tomb) bool {
	if command == "chpp" {

		fmt.Printf(" ^ WARNING: changing passphrase WILL BREAK dropbox sync. You have been warned!\n")

		var err error
		var bVer, bNew []byte
		newMatchesVer := false

		for !newMatchesVer {

			fmt.Printf(" ^ enter new passphrase.\n")
			fmt.Printf(" leeloo $ ")
			bNew, err = terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				panic(err)
			}

			if paranoia.CheckForExitOrQuit(&bNew) {
				fmt.Printf(" ...aborting\n")
				return true
			}

			fmt.Printf("\n ^ verify new passphrase.\n")
			fmt.Printf(" leeloo $ ")
			bVer, err = terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				panic(err)
			}

			if paranoia.CheckForExitOrQuit(&bVer) {
				fmt.Printf(" ...aborting\n")
				return true
			}

			if !paranoia.AreBytesEqual(&bNew, &bVer) {
				fmt.Printf("\n ^ verification does not match!\n")
			} else {
				newMatchesVer = true
			}

		}

		data.KillBytes(&bVer)

		data.ScatterPassphrase(&bNew)
		data.SealTomb(tomb)

		data.KillBytes(&bNew)

		fmt.Printf("\n ^ passphrase successfully changed.\n")

		return true
	}
	return false
}
