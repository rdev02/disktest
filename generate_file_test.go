package main

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestGenerateLen(t *testing.T) {
	res, err := GenerateLen(context.Background(), 1, "./a")
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
	if strings.Compare(expectedHash, res) != 0 {
		t.Errorf("expected %x, got %x", expectedHash, res)
	}
}
