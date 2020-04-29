package main

import (
	"errors"
	"fmt"
	"os"
)

type (
	inMemFile struct {
		file    *TempFile
		checked bool
	}

	//InMemRecorder holding records in memory
	InMemRecorder struct {
		filesMap map[string]*inMemFile
	}
)

//NewInMemRecorder constructor
func NewInMemRecorder() *InMemRecorder {
	return &InMemRecorder{
		filesMap: make(map[string]*inMemFile),
	}
}

func (rec InMemRecorder) RecordFile(file *TempFile) error {
	if file == nil {
		return errors.New("temp file can't be null")
	}

	if value, exist := rec.filesMap[file.hash]; exist {
		fmt.Fprintln(os.Stdout, "overwriting", file.hash, value.file.path, "->", file.path)
	}

	rec.filesMap[file.hash] = &inMemFile{
		file:    file,
		checked: false,
	}

	return nil
}

func (rec InMemRecorder) VerifyFileExits(file *TempFile) (bool, error) {
	if file == nil {
		return false, errors.New("temp file can't be null")
	}

	_, exists := rec.filesMap[file.hash]
	return exists, nil
}

func (rec InMemRecorder) MarkFileExits(file *TempFile) (bool, error) {
	if file == nil {
		return false, errors.New("temp file can't be null")
	}

	var err error = nil
	val, ok := rec.filesMap[file.hash]
	if ok {
		if val.checked {
			fmt.Println("WARN", file.hash, "has already been marked")
		}

		val.checked = true
	} else {
		err = fmt.Errorf("%s does not exist", file.hash)
	}

	return ok, err
}

func (rec InMemRecorder) FilesNotCheckedYet() ([]*TempFile, error) {
	result := make([]*TempFile, 0)
	for _, tmp := range rec.filesMap {
		if !tmp.checked {
			result = append(result, tmp.file)
		}
	}

	return result, nil
}
