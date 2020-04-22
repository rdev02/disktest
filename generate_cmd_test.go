package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	sizeFormat "github.com/rdev02/size-format"
)

func TestWriteVolume(t *testing.T) {
	errCh := make(chan error)
	workQ := make(chan (*TempFile))
	go func() {
		defer close(workQ)
		workQ <- &TempFile{path: "./a", size: sizeFormat.KB}
		workQ <- &TempFile{path: "./b", size: sizeFormat.KB}
	}()

	doneQ := make(chan (*TempFile))
	cnt := 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func(cnt *int) {
		defer wg.Done()
		fmt.Println("starting select")
	mainLoop:
		for {
			select {
			case proc, ok := <-doneQ:
				if !ok {
					break mainLoop
				}
				fmt.Println("processed", proc, *cnt)
				defer os.Remove(proc.path)
				*cnt++
				if *cnt == 2 {
					close(doneQ)
				}
			case err, ok := <-errCh:
				if !ok {
					break mainLoop
				}
				t.Error(err)
			}
		}

		fmt.Println("end verification function")
	}(&cnt)

	wg.Add(1)
	go writeVolume(context.Background(), workQ, doneQ, &wg, errCh)

	wg.Wait()

	if cnt != 2 {
		t.Error("expected 2 files, got", cnt)
	}
}

func TestWriteRandomFile(t *testing.T) {
	defer os.Remove("./a")
	tmpFile := TempFile{
		path: "./a",
		size: sizeFormat.KB,
	}

	err := writeRandomFile(context.Background(), &tmpFile)
	if err != nil {
		t.Error(err)
	}

	if len(tmpFile.hash) == 0 {
		t.Error("expected", tmpFile.hash, "to be populated")
	}

	tmpFile.path = "/non-xistent/h"
	tmpFile.hash = ""
	err = writeRandomFile(context.Background(), &tmpFile)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetRandomFileSizeFunc(t *testing.T) {
	constraint := tempFileSizeConstraint{
		min: 1 * sizeFormat.KB,
		max: 15 * sizeFormat.KB,
	}

	cap := int64(5 * sizeFormat.KB)

	f := getRandomFileSizeFunc(&constraint, cap)

	generated := f()
	if generated < constraint.min || generated > cap {
		t.Error("generated values should have been between", constraint.min, "and", cap, "but was", generated)
	}

	generated2 := f()
	if generated2 < constraint.min || generated2 > cap || generated == generated2 {
		t.Error("generated values should have been between", constraint.min, "and", cap, "and be different from", generated, "but was", generated2)
	}

	constraint.min = 100 * sizeFormat.KB
	constraint.max = 150 * sizeFormat.KB
	cap = 151 * sizeFormat.GB
	generated3 := getRandomFileSizeFunc(&constraint, cap)()
	if generated3 < constraint.min || generated3 > cap || generated > constraint.max {
		t.Error("generated values should have been between", constraint.min, "and", cap, "but was", generated3)
	}
}
