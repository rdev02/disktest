package main

import (
	"context"
	"testing"
)

//VerifyCmd start the generated fs verification process
func TestVerifyCmd(t *testing.T) {
	errQ := make(chan error)

	wg, err := VerifyCmd(context.Background(), nil, "./res", errQ)
	if err == nil {
		t.Error("Expected error on nil recorder")
	}

	recordingStrategy := IFileRecorder(NewInMemRecorder())
	go func() {
		defer close(errQ)

		wg, err = VerifyCmd(context.Background(), &recordingStrategy, "./res", errQ)
		if err != nil {
			t.Error(err)
		}

		wg.Wait()
	}()

	select {
	case err, ok := <-errQ:
		if !ok {
			break
		}
		t.Error(err)
	}
}

func TestVerifyVolume(t *testing.T) {
	recordingStrategy := IFileRecorder(NewInMemRecorder())
	errQ := make(chan error)
	go func() {
		defer close(errQ)
		verifyVolume(context.Background(), &recordingStrategy, "./res", errQ)
	}()

	select {
	case err, ok := <-errQ:
		if !ok {
			break
		}
		t.Error(err)
	}
}

func TestRecordVolume(t *testing.T) {
	doneQ := make(chan (*TempFile))
	errQ := make(chan error)
	recordingStrategy := IFileRecorder(NewInMemRecorder())

	go func() {
		defer close(doneQ)
		doneQ <- new(TempFile)
		doneQ <- new(TempFile)
	}()

	go func() {
		defer close(errQ)
		recordVolume(context.Background(), &recordingStrategy, doneQ, errQ)
	}()

	select {
	case err, ok := <-errQ:
		if !ok {
			break
		}
		t.Error(err)
	}
}