package data

import (
	"crypto/rand"
	"encoding/binary"
)

const (
	EntryFlagDeleted = (1 << 0)
)

type Entry struct {
	Clock uint32   `gob:"clk"`
	Key   string   `gob:"k"`
	Value string   `gob:"v"`
	Tags  []string `gob:"a"`
	Flags uint32   `gob:"f"`
}

type Tomb struct {
	ID         uint64  `gob:"id"`
	Version    uint64  `gob:"ver"`
	RandomPad0 []byte  `gob:"r0"`
	Entries    []Entry `gob:"es"`
	RandomPad1 []byte  `god:"r1"`
}

// --

var _Runes = []rune("qwertyuiopasdfghjkzxcvbnm23456789QWERTYUPASDFGHJKLZXCVBNM123456789")

func CreateCSPassword(n int) string {

	rlen := uint32(len(_Runes))

	if n > 32 {
		n = 32
	}

	entropy := make([]byte, 160, 160)
	nread, err := rand.Read(entropy)
	if err != nil {
		panic(err)
	}
	if nread != 160 {
		panic("did not read 160 bytes of entropy!")
	}

	b := make([]rune, n)
	for i := range b {
		idx := binary.LittleEndian.Uint32(entropy[i*4 : i*4+4])
		b[i] = _Runes[idx%rlen]
	}
	return string(b)
}
