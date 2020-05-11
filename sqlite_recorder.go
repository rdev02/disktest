package main

import (
	"errors"
)

type (
	SqliteRecorder struct {
	}
)

//NewSqlLiteRecorder constructor
func NewSqlLiteRecorder() *SqliteRecorder {
	return &SqliteRecorder{}
}

//RecordFile implements IFileRecorder
func (rec SqliteRecorder) RecordFile(file *TempFile) error {
	if file == nil {
		return errors.New("temp file can't be null")
	}

	//TODO: implement

	return nil
}

//VerifyFileExits implements IFileRecorder
func (rec SqliteRecorder) VerifyFileExits(file *TempFile) (bool, error) {
	if file == nil {
		return false, errors.New("temp file can't be null")
	}

	//TODO: implement
	return false, nil
}

//MarkFileExits implements IFileRecorder
func (rec SqliteRecorder) MarkFileExits(file *TempFile) (bool, error) {
	if file == nil {
		return false, errors.New("temp file can't be null")
	}

	//TODO implement
	return false, nil
}

//FilesNotCheckedYet implements IFileRecorder
func (rec SqliteRecorder) FilesNotCheckedYet() ([]*TempFile, error) {
	result := make([]*TempFile, 0)
	// TODO implement

	return result, nil
}

//GetTotalUnmarked implements IFileRecorder
func (rec SqliteRecorder) GetTotalUnmarked() (int64, error) {
	res := int64(0)

	return res, nil
}

//GetTotalMarked implements IFileRecorder
func (rec SqliteRecorder) GetTotalMarked() (int64, error) {
	res := int64(0)

	return res, nil
}
