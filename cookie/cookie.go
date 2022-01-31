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

func Load() string {
	cookiePath := path.Join(Directory, "cookie.txt")
	cookieFile, err := os.Open(cookiePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ""
		} else {
			log.Fatal(err)
		}
	}
	defer cookieFile.Close()

	b, err := ioutil.ReadAll(cookieFile)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}

func Save(cookie string) {
	err := os.MkdirAll(Directory, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	cookieFile, err := os.Create(path.Join(Directory, "cookie.txt"))
	if err != nil {
		log.Fatal(err)
	}
	defer cookieFile.Close()

	_, err = cookieFile.Write([]byte(cookie))
	if err != nil {
		log.Fatal(err)
	}
}

func Clear() {
	err := os.Remove(path.Join(Directory, "cookie.txt"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}
}
