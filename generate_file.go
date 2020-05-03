package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
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

	fileName := filepath.Base(path)

	hash := md5.New()
	hashedWriter := io.MultiWriter(f, hash)
	actualBuffer := size
	if size > defaultBuffer {
		actualBuffer = defaultBuffer
	}
	tmp := make([]byte, actualBuffer)

	rand.Seed(time.Now().UnixNano())
	var errorWrite error = nil
	allDone := make(chan int)
	var written int64 = 0
	go func() {
		defer close(allDone)

		for ; written < size; written += defaultBuffer {
			select {
			case <-ctx.Done():
				break
			default:
			}

			rand.Read(tmp)
			_, err := hashedWriter.Write(tmp)
			if err != nil {
				errorWrite = err
			}
		}
	}()

waitLoop:
	for {
		select {
		case <-allDone:
			break waitLoop
		case <-time.After(1 * time.Minute):
			fmt.Println(
				fmt.Sprintf("%s: [%s/%s] %.2f",
					fileName,
					sizeFormat.ToString(uint64(written)),
					sizeFormat.ToString(uint64(size)),
					math.Round((float64(written)/float64(size))*100)),
				"%")
		}
	}

	return fmt.Sprintf("%x", string(hash.Sum(nil))), errorWrite
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
