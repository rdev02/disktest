package main

import (
	"errors"
	"fmt"
	"os"
)

type (
	inMemFile struct {
		file   *TempFile
		marked bool
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
		file:   file,
		marked: false,
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
		if val.marked {
			fmt.Println("WARN", file.hash, "has already been marked")
		}

		val.marked = true
	} else {
		err = fmt.Errorf("%s does not exist", file.hash)
	}

	return ok, err
}

//FilesNotCheckedYet implements IFileRecorder
func (rec InMemRecorder) FilesNotCheckedYet() ([]*TempFile, error) {
	result := make([]*TempFile, 0)
	for _, tmp := range rec.filesMap {
		if !tmp.marked {
			result = append(result, tmp.file)
		}
	}

	return result, nil
}

//GetTotalUnmarked implements IFileRecorder
func (rec InMemRecorder) GetTotalUnmarked() (int64, error) {
	res := int64(0)
	for _, tmp := range rec.filesMap {
		if !tmp.marked {
			res += tmp.file.size
		}
	}

	return res, nil
}

//GetTotalMarked implements IFileRecorder
func (rec InMemRecorder) GetTotalMarked() (int64, error) {
	res := int64(0)
	for _, tmp := range rec.filesMap {
		if tmp.marked {
			res += tmp.file.size
		}
	}

	return res, nil
}
