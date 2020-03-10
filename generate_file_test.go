package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGenerateLen(t *testing.T) {
	res, err := generateLen(1, "./a")
	defer os.Remove("./a")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if len(res) <= 0 {
		t.Errorf("unexpected non empty string, got %x", res)
	}
}

func TestGetFileMd5(t *testing.T) {
	res, err := getFileMd5("./res/tst")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	expectedHash := string("f1341e91c533e8c0f79fa642e0151eb0")
	actualHash := fmt.Sprintf("%x", res)
	if strings.Compare(expectedHash, actualHash) != 0 {
		t.Errorf("expected %x, got %x", expectedHash, actualHash)
	}
}
