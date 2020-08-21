package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/caeril/leeloo/data"
	"github.com/caeril/leeloo/paranoia"
	"github.com/caeril/leeloo/shell"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {

	flagInit := flag.Bool("init", false, "initializes leeloo data file")
	flag.Parse()

	data.Init()

	if !data.IsTomb() {
		if *flagInit {
			for true {
				fmt.Printf(" ^ I will need a high-entropy passphrase.\n")
				fmt.Printf(" leeloo $ ")
				bpp0, _ := terminal.ReadPassword(int(syscall.Stdin))
				if len(bpp0) < 4 {
					fmt.Printf("\n ^ passphrase is of insufficient length.\n")
					continue
				}
				fmt.Printf("\n ^ ok, verify the passphrase.\n")
				fmt.Printf(" leeloo $ ")
				bpp1, _ := terminal.ReadPassword(int(syscall.Stdin))

				if !paranoia.AreBytesEqual(&bpp0, &bpp1) {
					fmt.Printf("\n ^ passphrase does not match verification.\n")
					continue
				}
				data.KillBytes(&bpp1)

				data.ScatterPassphrase(&bpp0)
				data.KillBytes(&bpp0)
				data.InitTomb()
				fmt.Printf(" ^ the leeloo's crypt is ready. tread if you dare.\n")

				data.Cleanup()
				os.Exit(0)
			}
		}
		fmt.Printf("leeloo has nothing to use. try leeloo -init\n")
		os.Exit(0)
	}

	if *flagInit {
		fmt.Printf("You already have a leeloo instance installed!\n")
		os.Exit(0)
	}

	fmt.Printf("\n l e e l o o   d a l l a s   m u l t i p a s s\n")

	var tomb data.Tomb

	for true {
		fmt.Printf("\n ^ enter your passphrase\n")
		fmt.Printf(" leeloo $ ")
		bpp, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			panic(err)
		}

		if paranoia.CheckForExitOrQuit(&bpp) {
			fmt.Printf("\n ^ goodbye.\n")
			data.Cleanup()
			os.Exit(0)
		}

		data.ScatterPassphrase(&bpp)
		data.KillBytes(&bpp)
		tomb, err = data.OpenTomb()

		if err == nil {
			break
		} else {
			if err == data.ErrVersionMismatch {
				continue
			} else {
				panic(err)
			}
		}
	}

	fmt.Printf("\n ^ opened crypt id %d version %d\n", tomb.ID, tomb.Version)

	stdin := bufio.NewReader(os.Stdin)
	var line string

	for true {

		fmt.Printf(" leeloo $ ")
		line, _ = stdin.ReadString('\n')
		if strings.HasSuffix(line, "\n") {
			line = line[0 : len(line)-1]
		}

		// tokenize!
		toks := strings.Split(line, " ")
		if len(toks) == 0 {
			continue
		}

		handled := shell.ProcessTokens(&tomb, toks)

		if !handled {
			fmt.Printf(" ^ %s ain't no command I ever heard of\n", toks[0])
		}

	}

}
