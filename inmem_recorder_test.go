package main

import (
	"strings"
	"testing"
)

func TestNewInMemRecorder(t *testing.T) {
	rec := NewInMemRecorder()
	if rec == nil {
		t.Error("expected non nil object")
	}
}

func TestRecordFile(t *testing.T) {
	rec := NewInMemRecorder()

	f1 := TempFile{
		hash: "hash1",
	}

	rec.RecordFile(&f1)

	if len(rec.filesMap) != 1 {
		t.Error("expected internal map len to be 1, instead, saw", len(rec.filesMap))
	}

	if _, ok := rec.filesMap[f1.hash]; !ok {
		t.Error("expected value to be present in the map", f1)
	}

	rec.RecordFile(&f1)

	if len(rec.filesMap) != 1 {
		t.Error("expected internal map len to be 1, instead, saw", len(rec.filesMap))
	}

	f2 := TempFile{
		hash: "hash2",
	}

	rec.RecordFile(&f2)

	if len(rec.filesMap) != 2 {
		t.Error("expected internal map len to be 2, instead, saw", len(rec.filesMap))
	}

	if _, ok := rec.filesMap[f2.hash]; !ok {
		t.Error("expected value to be present in the map", f2)
	}
}

func TestVerifyFileExits(t *testing.T) {
	rec := NewInMemRecorder()

	f1 := TempFile{
		hash: "hash1",
	}
	f2 := TempFile{
		hash: "hash2",
	}

	rec.RecordFile(&f1)
	rec.RecordFile(&f2)

	if rec, ok := rec.VerifyFileExits(&f1); !rec || ok != nil {
		t.Error("unexpected", rec, ok)
	}

	if rec, ok := rec.VerifyFileExits(&f2); !rec || ok != nil {
		t.Error("unexpected", rec, ok)
	}

	if rec, ok := rec.VerifyFileExits(nil); rec || ok == nil {
		t.Error("unexpected", rec, ok)
	}

	f3 := TempFile{
		hash: "hash3",
	}

	if rec, ok := rec.VerifyFileExits(&f3); rec || ok != nil {
		t.Error("unexpected", rec, ok)
	}
}

func TestMarkFileExits(t *testing.T) {
	rec := NewInMemRecorder()

	f1 := TempFile{
		hash: "hash1",
	}
	f2 := TempFile{
		hash: "hash2",
	}

	rec.RecordFile(&f1)

	if marked, err := rec.MarkFileExits(&f1); !marked || err != nil {
		t.Error("unexpected", rec, err)
	}

	if marked, err := rec.MarkFileExits(nil); marked || err == nil {
		t.Error("unexpected", rec, err)
	}

	if marked, err := rec.MarkFileExits(&f2); marked || err == nil {
		t.Error("unexpected", rec, err)
	}
}

func TestFilesNotCheckedYet(t *testing.T) {
	rec := NewInMemRecorder()

	f1 := TempFile{
		hash: "hash1",
	}
	f2 := TempFile{
		hash: "hash2",
	}

	rec.RecordFile(&f1)
	rec.RecordFile(&f2)
	rec.MarkFileExits(&f1)

	notChecked, err := rec.FilesNotCheckedYet()
	if err != nil || len(notChecked) != 1 {
		t.Error("unexpected", err, notChecked)
	}

	if strings.Compare(notChecked[0].hash, f2.hash) != 0 {
		t.Error("unexpected", notChecked[0].hash, f2.hash)
	}

}
