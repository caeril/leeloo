package boojum

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"
)

func CreateNaiveBoojum(f ErrorHandler) Boojum {
	b := NaiveBoojum{
		errorHandler: f,
		running:      true,
		stopped:      true,
	}

	b.init()

	go func() {
		b.stopped = false
		for b.running {
			// 750ms shuffle, a little less than Schneier's "at least" suggestion
			time.Sleep(time.Millisecond * 750)
			b.shuffle()
		}
		b.stopped = true
	}()
	return &b
}

//

func fillRandom(ba *[]byte) error {
	n, err := rand.Read(*ba)
	if err != nil {
		return err
	}
	if n != len(*ba) {
		return errors.New(fmt.Sprintf("Could only read %d bytes", n))
	}
	return nil
}

/*
	randomSpace[128:144] is the iv
	randomSpace[256:?] is the encrypted key material
*/

const MAX_LENGTH_BYTES = 1024
const RANDOM_SPACE_LENGTH = 1024 * 2048
const I_LOCATION = 1031 * 281  // location of xor'ed i
const R_LOCATION = 1031 * 1279 // location of first to hash
const S_LOCATION = 1031 * 1987 // location of second to hash

type NaiveBoojum struct {
	running      bool // set to false to end shuffle loop
	stopped      bool // naming things is hard! this gets set to true when shuffle loop actually ends
	randomSpace  []byte
	lock         *sync.Mutex
	dataLength   uint32
	errorHandler ErrorHandler
}

func (b *NaiveBoojum) init() {

	b.randomSpace = make([]byte, RANDOM_SPACE_LENGTH, RANDOM_SPACE_LENGTH)
	b.lock = &sync.Mutex{}

	err := fillRandom(&b.randomSpace)
	if err != nil {
		panic(err)
	}

}

func (b *NaiveBoojum) geti() []byte {
	i := make([]byte, 32, 32)
	rHash := sha256.Sum256(b.randomSpace[R_LOCATION : R_LOCATION+1024])
	sHash := sha256.Sum256(b.randomSpace[S_LOCATION : S_LOCATION+1024])
	copy(i, b.randomSpace[I_LOCATION:I_LOCATION+32])

	for j := 0; j < 32; j++ {
		i[j] ^= rHash[j]
	}
	for j := 0; j < 32; j++ {
		i[j] ^= sHash[j]
	}

	return i
}

func (b *NaiveBoojum) shuffle() {

	b.lock.Lock()

	// grab raw key material
	raw, err := b.get(true)
	if err != nil {
		b.lock.Unlock()
		b.errorHandler(err)
		return
	}

	// kill old random space
	err = fillRandom(&b.randomSpace)
	if err != nil {
		b.lock.Unlock()
		b.errorHandler(err)
		return
	}

	// create and fill new random space
	b.randomSpace = make([]byte, RANDOM_SPACE_LENGTH, RANDOM_SPACE_LENGTH)
	err = fillRandom(&b.randomSpace)
	if err != nil {
		b.lock.Unlock()
		b.errorHandler(err)
		return
	}

	// re-set and kill raw slice
	err = b.set(raw, true)
	if err != nil {
		b.lock.Unlock()
		b.errorHandler(err)
		return
	}
	err = fillRandom(&raw)
	if err != nil {
		b.lock.Unlock()
		b.errorHandler(err)
		return
	}

	b.lock.Unlock()

}

func (b *NaiveBoojum) Set(ba []byte) error {
	return b.set(ba, false)
}

func (b *NaiveBoojum) set(ba []byte, locked bool) error {

	if len(ba) > MAX_LENGTH_BYTES {
		return errors.New(fmt.Sprintf("data length exceeds maximum length of %d bytes", len(ba)))
	}

	if !locked {
		b.lock.Lock()
	}
	crypted := make([]byte, MAX_LENGTH_BYTES, MAX_LENGTH_BYTES)
	key := b.geti()
	algo, err := aes.NewCipher(key)
	if err != nil {
		fillRandom(&key)
		if !locked {
			b.lock.Unlock()
		}
		return err
	}

	tocrypt := make([]byte, MAX_LENGTH_BYTES, MAX_LENGTH_BYTES)
	copy(tocrypt[0:len(ba)], ba)

	stream := cipher.NewCTR(algo, b.randomSpace[128:144])
	stream.XORKeyStream(crypted, tocrypt)

	// kill key
	err = fillRandom(&key)
	if err != nil {
		if !locked {
			b.lock.Unlock()
		}
		return err
	}

	// kill tocrypt
	err = fillRandom(&tocrypt)
	if err != nil {
		fillRandom(&key)
		if !locked {
			b.lock.Unlock()
		}
		return err
	}

	// copy back to randomSpace
	copy(b.randomSpace[256:256+MAX_LENGTH_BYTES], crypted)

	// set length, kill input, and unlock
	b.dataLength = uint32(len(ba))
	err = fillRandom(&ba)
	if err != nil {
		fillRandom(&key)
		if !locked {
			b.lock.Unlock()
		}
		return err
	}

	if !locked {
		b.lock.Unlock()
	}
	return nil

}

func (b *NaiveBoojum) Get() ([]byte, error) {
	return b.get(false)
}

func (b *NaiveBoojum) get(locked bool) ([]byte, error) {
	ba := make([]byte, b.dataLength, b.dataLength)

	if !locked {
		b.lock.Lock()
	}

	key := b.geti()
	algo, err := aes.NewCipher(key)
	if err != nil {
		fillRandom(&key)
		if !locked {
			b.lock.Unlock()
		}
		return nil, err
	}

	decrypted := make([]byte, MAX_LENGTH_BYTES, MAX_LENGTH_BYTES)

	stream := cipher.NewCTR(algo, b.randomSpace[128:144])
	stream.XORKeyStream(decrypted, b.randomSpace[256:256+MAX_LENGTH_BYTES])

	// kill key
	err = fillRandom(&key)
	if err != nil {
		if !locked {
			b.lock.Unlock()
		}
		return nil, err
	}

	copy(ba, decrypted[0:b.dataLength])

	// kill decrypted bytes
	err = fillRandom(&decrypted)
	if err != nil {
		if !locked {
			b.lock.Unlock()
		}
		return nil, err
	}

	if !locked {
		b.lock.Unlock()
	}
	return ba, nil
}

func (b *NaiveBoojum) Cleanup() error {

	b.running = false // hope to Zeus that this is atomic!
	for !b.stopped {
		time.Sleep(time.Millisecond * 10)
	}

	// final kill of randomSpace
	err := fillRandom(&b.randomSpace)
	if err != nil {
		return err
	}

	return nil
}
