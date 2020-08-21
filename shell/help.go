package shell

import "fmt"

func checkHelp(command string, tokens []string) bool {
	if command == "help" {
		fmt.Printf(" ^ help: \t\tthis menu\n")
		fmt.Printf(" ^ list <arg>: \t\tlists all keys matching <arg>\n")
		fmt.Printf(" ^ show <arg>: \t\tshows all key/values matching <arg> in key\n")
		fmt.Printf(" ^ add <key> <value>: \tadds new entry\n")
		fmt.Printf(" ^ gen <key> <len>: \tadds new entry, ewith generated value of length <len>\n")
		fmt.Printf(" ^ del <key>: \t\tdeletes entry\n")
		fmt.Printf(" ^ sync: \t\tsynchronizes to (dropbox) origin\n")
		fmt.Printf(" ^ chpp:  \t\tchanges passphrase\n")
		return true
	}
	return false
}
