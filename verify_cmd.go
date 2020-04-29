package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//VerifyCmd start the generated fs verification process
func VerifyCmd(ctx context.Context, recorder *IFileRecorder, volumeRoot string, errorChan chan<- error) (*sync.WaitGroup, error) {
	if recorder == nil {
		return nil, errors.New("recorder can't be nil")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		verifyVolume(ctx, recorder, volumeRoot, errorChan)
	}()

	return &wg, nil
}

func verifyVolume(ctx context.Context, recorder *IFileRecorder, volumeRoot string, errorChan chan<- error) {
	rec := *recorder

	// now verify we can read back all we wrote
	filepath.Walk(volumeRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading", path, err)
			errorChan <- err
			return err
		}

		select {
		case _, ok := <-ctx.Done():
			fmt.Println("cancelling file walk", ok)
			return errors.New("context cancel")
		default:
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				// skip hidden folders: we do not generated them
				return filepath.SkipDir
			}

			// we are only concerned with files
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			// skip hidden files
			return nil
		}

		fileHash, err := GetFileMd5(path)
		if err != nil {
			errorChan <- err
			return err
		}

		file := TempFile{
			path: path,
			size: info.Size(),
			hash: fileHash,
		}

		if ok, err := rec.VerifyFileExits(&file); !ok || err != nil {
			fmt.Fprintln(os.Stdout, "WARN: file", path, file.hash, "was not recorded previously", err)
			return err
		}

		_, err = rec.MarkFileExits(&file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERR: could not mark file as existing ", path, file.hash)
			errorChan <- err
			return err
		}

		return nil
	})

	remainingFiles, err := rec.FilesNotCheckedYet()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERR: could not get missing files", err)
		return
	}

	if len(remainingFiles) > 0 {
		fmt.Fprintln(os.Stderr, "ERR: not all files were read/verified. Missing files:")
		for _, file := range remainingFiles {
			fmt.Fprintln(os.Stderr, file.path)
		}
	} else {
		fmt.Println("Success: all files were read and verified")
	}
}

func recordVolume(ctx context.Context, recorder *IFileRecorder, doneQueue <-chan (*TempFile), errorChan chan<- error) {
	rec := *recorder

	for workItem := range processOrDone(ctx, doneQueue) {
		err := rec.RecordFile(workItem)
		if err != nil {
			errorChan <- err
			break
		}
	}
}
