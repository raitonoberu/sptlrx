package cookie

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var Directory string

func init() {
	d, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	Directory = path.Join(d, "sptlrx")
}

func Load() (string, error) {
	cookiePath := path.Join(Directory, "cookie.txt")
	cookieFile, err := os.Open(cookiePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		} else {
			return "", err
		}
	}
	defer cookieFile.Close()

	b, err := ioutil.ReadAll(cookieFile)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func Save(cookie string) error {
	err := os.MkdirAll(Directory, os.ModePerm)
	if err != nil {
		return err
	}

	cookieFile, err := os.Create(path.Join(Directory, "cookie.txt"))
	if err != nil {
		return err
	}
	defer cookieFile.Close()

	_, err = cookieFile.Write([]byte(cookie))
	return err
}

func Clear() error {
	return os.Remove(path.Join(Directory, "cookie.txt"))
}
