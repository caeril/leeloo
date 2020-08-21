package data

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/caeril/leeloo/boojum"
	"github.com/caeril/leeloo/paranoia"
	"golang.org/x/crypto/scrypt"
)

var booj boojum.Boojum

/*
scrypt params for reference:
func Key(password, salt []byte, N, r, p, keyLen int) ([]byte, error)
*/

func Init() {
	booj = boojum.Init(func(err error) {
		fmt.Printf(" ^ boojum init has failed: %s\n", err.Error())
		os.Exit(0)
	})
}

func Cleanup() {
	booj.Cleanup()
}

func ScatterPassphrase(bpp *[]byte) {

	naiveHashSha := sha256.Sum256(*bpp)

	naiveHashSalt := make([]byte, 32, 32)
	paranoia.PanicedRandomRead(&naiveHashSalt)

	kh, err := scrypt.Key(naiveHashSha[:], naiveHashSalt, 32768, 16, 1, 32)
	if err != nil {
		panic(err)
	}
	naiveHashScrypt := sha256.Sum256(kh)

	construct := make([]byte, 128)

	copy(construct[0:32], naiveHashSha[:])
	copy(construct[32:64], naiveHashScrypt[:])
	copy(construct[64:96], naiveHashSalt)

	paranoia.PanicedRandomRead(&naiveHashSalt)
	paranoia.PanicedRandomRead32(&naiveHashScrypt)
	paranoia.PanicedRandomRead32(&naiveHashSha)

	booj.Set(construct)

	paranoia.PanicedRandomRead(&construct)

}

func KillBytes(ba *[]byte) {
	n, err := rand.Read(*ba)
	if err != nil {
		panic(err)
	}
	if n != len(*ba) {
		panic("failed to kill all bytes in slice")
	}
}

// internal stuff

func gatherHashSha() *[32]byte {
	out := [32]byte{}
	construct, err := booj.Get()
	if err != nil {
		panic(err)
	}
	copy(out[:], construct[0:32])
	paranoia.PanicedRandomRead(&construct)
	return &out
}

func gatherHashScrypt() *[32]byte {
	out := [32]byte{}
	construct, err := booj.Get()
	if err != nil {
		panic(err)
	}
	copy(out[:], construct[32:64])
	paranoia.PanicedRandomRead(&construct)
	return &out
}

func gatherHashScryptWithSalt(salt []byte) *[32]byte {

	naiveHashSha := gatherHashSha()
	kh, err := scrypt.Key(naiveHashSha[:], salt, 32768, 16, 1, 32)
	paranoia.PanicedRandomRead32(naiveHashSha)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(kh)

	out := [32]byte{}
	copy(out[:], hash[:])
	paranoia.PanicedRandomRead32(&hash)
	return &out
}

func gatherHashSalt() *[]byte {
	out := make([]byte, 32, 32)
	construct, err := booj.Get()
	if err != nil {
		panic(err)
	}
	copy(out, construct[64:96])
	paranoia.PanicedRandomRead(&construct)
	return &out

}

func killHash(hash *[32]byte) {
	paranoia.PanicedRandomRead32(hash)
}
