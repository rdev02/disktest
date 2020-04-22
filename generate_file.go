package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	sizeFormat "github.com/rdev02/size-format"
)

const defaultBuffer = 20 * sizeFormat.MB

//GenerateLen generates a file of size at path, returns MD5 hash.
func GenerateLen(ctx context.Context, size int64, path string) (string, error) {
	if size <= 0 {
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
	var written int64 = 0
	for ; written < size; written += defaultBuffer {
		select {
		case <-ctx.Done():
			break
		default:
		}

		rand.Read(tmp)
		_, err := hashedWriter.Write(tmp)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", string(hash.Sum(nil))), nil
}

//GetFileMd5 generates MD5 of the file at path
func GetFileMd5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", string(h.Sum(nil))), nil
}
