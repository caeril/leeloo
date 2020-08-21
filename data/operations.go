package data

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/caeril/leeloo/paranoia"
	"golang.org/x/crypto/nacl/secretbox"
)

var _tombDir = ""
var _tombPath = ""
var ErrVersionMismatch = errors.New("version mismatch")

func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	_tombDir = usr.HomeDir + "/" + ".leeloo"
	_tombPath = _tombDir + "/current.dat"
}

func IsTomb() bool {
	_, err := os.Stat(_tombPath)
	return !os.IsNotExist(err)
}

func InitTomb() {

	entropy := make([]byte, 8, 8)
	rand.Read(entropy)

	// create initial tomb
	tomb := Tomb{}
	tomb.Version = 1
	tomb.ID = binary.LittleEndian.Uint64(entropy)
	tomb.Entries = append(tomb.Entries, Entry{Key: "leeloo.first.data", Value: "poop"})

	SealTomb(&tomb)

}

// file format:
// [4 byte preamble][8 byte id][8 byte version][32 byte nonce][32 byte salt - scrypt format only!][data]
// preamble 0x18e9b0a5 :: old sha256 key derivation. Sad!
// preamble 0x18e9b0a6 :: new scrypt key derivation
const dataOffsetSha = 52
const dataOffsetScrypt = 84

func SealTomb(tomb *Tomb) {

	entropy := make([]byte, 128, 128)
	n, err := rand.Read(entropy)
	if err != nil {
		panic(err)
	}
	if n != 128 {
		panic("did not read in 128 bytes of entropy!")
	}

	tomb.Version++

	// find a decent nonce
	nonce := entropy[64:96]

	// fill tomb with some random padding - make exact length difficult to guess
	l0 := (paranoia.CSRandomUint32() % 2048) + 128
	l1 := (paranoia.CSRandomUint32() % 2048) + 128
	tomb.RandomPad0 = make([]byte, l0, l0)
	tomb.RandomPad1 = make([]byte, l1, l1)
	rand.Read(tomb.RandomPad0)
	rand.Read(tomb.RandomPad1)

	nonce24 := [24]byte{}
	copy(nonce24[:], nonce[4:28])

	// encode
	encodedTombBuffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&encodedTombBuffer)
	encoder.Encode(&tomb)
	encodedTomb := encodedTombBuffer.Bytes()

	// Encrypt
	hash := gatherHashScrypt()
	encryptedTomb := secretbox.Seal([]byte{}, encodedTomb, &nonce24, hash)
	killHash(hash)

	// get current salt
	salt := gatherHashSalt()

	file := make([]byte, dataOffsetScrypt+len(encryptedTomb), dataOffsetScrypt+len(encryptedTomb))
	binary.LittleEndian.PutUint32(file[0:4], 0x18e9b0a6)
	binary.LittleEndian.PutUint64(file[4:12], tomb.ID)
	binary.LittleEndian.PutUint64(file[12:20], tomb.Version)
	copy(file[20:52], nonce)
	copy(file[52:84], *salt)
	copy(file[84:], encryptedTomb)

	KillBytes(salt)

	SealThisTomb(file)

}

func SealThisTomb(dataz []byte) error {

	// verify path
	_, err := os.Stat(_tombDir)
	if os.IsNotExist(err) {
		os.Mkdir(_tombDir, os.FileMode(0700))
	}

	err = ioutil.WriteFile(_tombPath, dataz, os.FileMode(0600))
	return err
}

func GetSealedTomb() []byte {
	file, err := ioutil.ReadFile(_tombPath)
	if err != nil {
		panic(err)
	}
	return file
}

func OpenTomb() (Tomb, error) {

	file, err := ioutil.ReadFile(_tombPath)
	if err != nil {
		panic(err)
	}
	return OpenThisTomb(file)
}

func OpenThisTomb(file []byte) (Tomb, error) {

	useSha := false
	useScrypt := false
	fileStartIndex := 0
	preamble := binary.LittleEndian.Uint32(file[0:4])
	if preamble == 0x18e9b0a5 {
		useSha = true
		fileStartIndex = 52
	} else if preamble == 0x18e9b0a6 {
		useScrypt = true
		fileStartIndex = 84
	} else {
		panic("preamble does not match!")
	}
	id := binary.LittleEndian.Uint64(file[4:12])
	version := binary.LittleEndian.Uint64(file[12:20])
	nonce := make([]byte, 32, 32)
	copy(nonce, file[20:52])
	nonce24 := [24]byte{}
	copy(nonce24[:], nonce[4:28])

	tomb := Tomb{}

	// Decrypt
	var hash *[32]byte

	if useSha {
		hash = gatherHashSha()
	}
	if useScrypt {
		salt := make([]byte, 32, 32)
		copy(salt, file[52:84])
		hash = gatherHashScryptWithSalt(salt)
	}
	encodedTomb, _ := secretbox.Open([]byte{}, file[fileStartIndex:], &nonce24, hash)
	killHash(hash)

	encodedTombBuffer := bytes.NewBuffer(encodedTomb)
	decoder := gob.NewDecoder(encodedTombBuffer)
	decoder.Decode(&tomb)

	// Validate id and version
	if id != tomb.ID || version != tomb.Version {
		return tomb, ErrVersionMismatch
	}

	return tomb, nil
}
