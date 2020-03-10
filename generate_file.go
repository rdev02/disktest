package main

import (
	"crypto/md5"
	"errors"
	"io"
	"math/rand"
	"os"
	"time"

	sizeFormat "github.com/rdev02/size-format"
)

const defaultBuffer = 20 * sizeFormat.MB

func generateLen(size uint64, path string) (string, error) {
	if size == 0 {
		return "", errors.New("size must be greater then 0")
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	hashedWriter := io.MultiWriter(f, hash)
	actualBuffer := size
	if size > defaultBuffer {
		actualBuffer = defaultBuffer
	}
	tmp := make([]byte, actualBuffer)

	rand.Seed(time.Now().UnixNano())
	var written uint64 = 0
	for ; written < size; written += defaultBuffer {
		rand.Read(tmp)
		_, err := hashedWriter.Write(tmp)
		if err != nil {
			return "", err
		}
	}

	return string(hash.Sum(nil)), nil
}
