package main

import (
	"context"
	"os"
	"testing"

	sizeFormat "github.com/rdev02/size-format"
)

func BenchmarkGenerateCmd(b *testing.B) {
	rootPath := "build/test"
	size := 10 * sizeFormat.TB
	errCh := make(chan error)
	recorder := IFileRecorder(NewInMemRecorder())

	os.MkdirAll(rootPath, 0700)
	defer os.RemoveAll(rootPath)

	done := GenerateCmd(context.Background(), rootPath, int64(size), &recorder, errCh, pseudoWriteFile)
	done.Wait()
}
