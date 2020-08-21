package paranoia

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

//
// CSRandomUint32 is a convenience function for getting uint32s out of crypto/rand
//
func CSRandomUint32() uint32 {
	entropy := make([]byte, 8, 8)
	n, err := rand.Read(entropy)
	if err != nil {
		panic(err)
	}
	if n != len(entropy) {
		panic("failed to read 8 bytes of entropy")
	}
	return binary.BigEndian.Uint32(entropy[2:6])
}

//
// AreBytesEqual is a pathetic attempt to prevent timing attacks against
// byte array comparisons. we try to scale random delays by length of arrays
//
func AreBytesEqual(a *[]byte, b *[]byte) bool {
	delay1 := int(CSRandomUint32()) % (32 + len(*a)/2048)
	delay2 := int(CSRandomUint32()) % (32 + len(*b)/2048)
	time.Sleep(time.Millisecond * time.Duration(delay1))
	out := true
	if len(*a) != len(*b) {
		out = false
	} else {
		for i := range *a {
			if (*a)[i] != (*b)[i] {
				out = false
			}
		}
	}
	time.Sleep(time.Millisecond * time.Duration(delay2))
	return out
}

func PanicedRandomRead(ba *[]byte) {
	n, err := rand.Read(*ba)
	if err != nil {
		panic(err)
	}
	if n != len(*ba) {
		panic("did not kill em all!")
	}
}

func PanicedRandomRead32(ba *[32]byte) {
	tba := make([]byte, 32, 32)
	n, err := rand.Read(tba)
	if err != nil {
		panic(err)
	}
	if n != len(*ba) {
		panic("did not kill em all!")
	}
	for i := range tba {
		ba[i] = tba[i]
	}
}

func CheckForExitOrQuit(bpp *[]byte) bool {
	bexit := []byte("exit")
	bquit := []byte("quit")

	exit := AreBytesEqual(bpp, &bexit)
	quit := AreBytesEqual(bpp, &bquit)

	return exit || quit
}
