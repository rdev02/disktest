package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//VerifyCmd start the generated fs verification process
func VerifyCmd(ctx context.Context, recorder *IFileRecorder, volumeRoot string, errorChan chan<- error) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		verifyVolume(ctx, recorder, volumeRoot, errorChan)
	}()

	return &wg
}

func verifyVolume(ctx context.Context, recorder *IFileRecorder, volumeRoot string, errorChan chan<- error) {
	rec := *recorder

	// now verify we can read back all we wrote
	filepath.Walk(volumeRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading", path, err)
			_, cancel := context.WithCancel(ctx)
			cancel()
			return err
		}

		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			// skip hidden folders: we do not generated them
			return filepath.SkipDir
		}

		if strings.HasPrefix(info.Name(), ".") {
			// skip hidden files
			return nil
		}

		fileHash, err := GetFileMd5(path)
		if err != nil {

		}

		file := TempFile{
			path: path,
			size: uint64(info.Size()),
			hash: fileHash,
		}

		if ok, err := rec.VerifyFileExits(&file); !ok || err != nil {
			fmt.Fprintln(os.Stdout, "WARN: file", path, file.hash, "was not recorded previously", err)
			return err
		}

		_, err = rec.MarkFileExits(&file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERR: could not mark file as existing ", path, file.hash)
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

func recordVolume(ctx context.Context, recorder *IFileRecorder, doneQueue <-chan (*TempFile), result chan<- error) {
	rec := *recorder

	// add Volume recording as a unit of work
	defer close(result)

	for workItem := range processOrDone(ctx, doneQueue) {
		err := rec.RecordFile(workItem)
		if err != nil {
			result <- err
			break
		}
	}
}
