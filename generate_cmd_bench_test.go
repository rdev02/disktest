package main

import (
	"context"
	"os"
	"sync"
	"testing"

	sizeFormat "github.com/rdev02/size-format"
)

func BenchmarkGenerateCmd(b *testing.B) {
	rootPath := "build/test"
	size := 10 * sizeFormat.MB
	errCh := make(chan error)
	recorder := IFileRecorder(NewInMemRecorder())

	os.MkdirAll(rootPath, 0700)
	defer os.RemoveAll(rootPath)

	done := GenerateCmd(context.Background(), rootPath, int64(size), &recorder, errCh, pseudoWrittenFile)
	done.Wait()
}

func pseudoWrittenFile(ctx context.Context, workQueue <-chan (*TempFile), doneQueue chan<- (*TempFile), wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	hashLen := 10
	for workItem := range processOrDone(ctx, workQueue) {
		lenPath := len(workItem.path)
		if lenPath >= hashLen {
			workItem.hash = workItem.path[lenPath-hashLen:]
		} else {
			workItem.hash = workItem.path
		}

		doneQueue <- workItem
	}
}
