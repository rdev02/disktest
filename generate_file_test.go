package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestGenerateLen(t *testing.T) {
	res, err := GenerateLen(1, "./a")
	defer os.Remove("./a")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if len(res) <= 0 {
		t.Errorf("unexpected non empty string, got %x", res)
	}
}

func TestGetFileMd5(t *testing.T) {
	res, err := GetFileMd5("./res/tst")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	expectedHash := string("f1341e91c533e8c0f79fa642e0151eb0")
	actualHash := fmt.Sprintf("%x", res)
	if strings.Compare(expectedHash, actualHash) != 0 {
		t.Errorf("expected %x, got %x", expectedHash, actualHash)
	}
}

func TestCancelContextForAll(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan (rune))

	var wg sync.WaitGroup
	wg.Add(1)

	go goTest(ctx, ch, "1", &wg)
	go goTest(ctx, ch, "2", &wg)
	go goTest(ctx, ch, "3", &wg)
	go goTest(ctx, ch, "4", &wg)

	go func() {
		defer wg.Done()
		ch <- 11
		ch <- 12
		ch <- 13
		ch <- 14
		close(ch)
	}()

	fmt.Println("all done")
	wg.Wait()
	cancel()
}

func goTest(ctx context.Context, data chan (rune), num string, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	select {
	case <-ctx.Done():
		fmt.Println(num, "Exiting because of context")
		return
	case d, ok := <-data:
		fmt.Println(num, "got data", d, ok)
	}
}
