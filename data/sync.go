package data

// initial sync will only use dropbox api

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	dropbox "github.com/tj/go-dropbox"
)

var ErrInvalidDropboxToken = errors.New("invalid dropbox token, sorry")
var ErrDropboxNotFound = errors.New("not found on dropbox remote")

func GetRemote(id uint64) ([]byte, error) {

	token := os.Getenv("LEELOO_DROPBOX_TOKEN")
	path := fmt.Sprintf("/leeloo-%d.dat", id)

	dbx := dropbox.New(dropbox.NewConfig(token))
	resp, err := dbx.Files.Download(&dropbox.DownloadInput{Path: path})
	if err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return nil, ErrDropboxNotFound
		}
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.Length != int64(len(data)) {
		return nil, errors.New("confirmed size does not match received size!")
	}

	return data, nil

}

func PutRemote(id uint64, data []byte) error {
	token := os.Getenv("LEELOO_DROPBOX_TOKEN")
	path := fmt.Sprintf("/leeloo-%d.dat", id)

	dbx := dropbox.New(dropbox.NewConfig(token))
	resp, err := dbx.Files.Upload(&dropbox.UploadInput{Path: path,
		Reader: bytes.NewReader(data), Mode: dropbox.WriteModeOverwrite})
	if err != nil {
		return err
	}
	if resp.Size != uint64(len(data)) {
		return errors.New("confirmed size does not match sent size!")
	}

	return nil
}
