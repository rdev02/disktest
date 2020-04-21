package main

import (
	"errors"
)

type (
	SqliteRecorder struct {
	}
)

func NewSqlLiteRecorder() *SqliteRecorder {
	return &SqliteRecorder{}
}

func (rec SqliteRecorder) RecordFile(file *TempFile) error {
	if file == nil {
		return errors.New("temp file can't be null")
	}

	//TODO: implement

	return nil
}

func (rec SqliteRecorder) VerifyFileExits(file *TempFile) (bool, error) {
	if file == nil {
		return false, errors.New("temp file can't be null")
	}

	//TODO: implement
	return false, nil
}

func (rec SqliteRecorder) MarkFileExits(file *TempFile) (bool, error) {
	if file == nil {
		return false, errors.New("temp file can't be null")
	}

	//TODO implement
	return false, nil
}

func (rec SqliteRecorder) FilesNotCheckedYet() ([]*TempFile, error) {
	result := make([]*TempFile, 0)
	// TODO implement

	return result, nil
}
