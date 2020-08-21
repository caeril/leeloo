package shell

import (
	"fmt"

	"github.com/caeril/leeloo/data"
)

func checkSync(command string, tokens []string, tomb *data.Tomb) bool {

	if command == "sync" {
		fmt.Printf(" ^ synchronizing with origin...")

		dataz, err := data.GetRemote(tomb.ID)
		if err != nil {
			if err == data.ErrDropboxNotFound {
				fmt.Printf("( not found )...")
				err = data.PutRemote(tomb.ID, data.GetSealedTomb())
				if err != nil {
					fmt.Printf("ERROR [%s]!\n", err)
				} else {
					fmt.Printf("done!\n")
				}
				return true
			} else {
				fmt.Printf("ERROR RETRIEVING %s\n", err)
				return true
			}
		}

		// open remote tomb
		remoteTomb, err := data.OpenThisTomb(dataz)
		if err != nil {
			fmt.Printf("ERROR [%s]!\n", err)
		}

		// diff remote additions with local state - VERY NAIVE!
		// todo: implement better diff without spraying plaintext all over the place
		// todo: also make this transactional/atomic

		localMap := make(map[string][]data.Entry)
		for i := range tomb.Entries {
			localEntries, haz := localMap[tomb.Entries[i].Key+":"+tomb.Entries[i].Value]
			if !haz {
				localMap[tomb.Entries[i].Key+":"+tomb.Entries[i].Value] = append([]data.Entry{}, tomb.Entries[i])
			} else {
				localMap[tomb.Entries[i].Key+":"+tomb.Entries[i].Value] = append(localEntries, tomb.Entries[i])
			}
		}

		added := 0
		for i := range remoteTomb.Entries {
			_, haz := localMap[remoteTomb.Entries[i].Key+":"+remoteTomb.Entries[i].Value]
			if !haz {
				tomb.Entries = append(tomb.Entries, remoteTomb.Entries[i])
				added++
			}
		}

		data.SealTomb(tomb)

		err = data.PutRemote(tomb.ID, data.GetSealedTomb())
		if err != nil {
			fmt.Printf("error [%s]!\n", err)
			return true
		}

		fmt.Printf("done! (%d new entries)\n", added)

		return true

	}

	return false

}
