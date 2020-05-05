package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	sizeFormat "github.com/rdev02/size-format"
)

type (
	volumePathFolder struct {
		basePath string
		filesNum rune
	}

	tempFileSizeConstraint struct {
		min, max int64
	}

	writeFunc = func(ctx context.Context, workQueue <-chan (*TempFile), doneQueue chan<- (*TempFile), wg *sync.WaitGroup, errChan chan<- error)
)

const (
	numFilesPerFolder  = 500
	numSubfolders      = 10
	maxLargeFilesShare = .5
	maxMedFilesShare   = .35
)

var (
	smallFileSizeConstraint = tempFileSizeConstraint{min: 100 * sizeFormat.KB, max: 50 * sizeFormat.MB}
	medFileSizeConstraint   = tempFileSizeConstraint{min: 100 * sizeFormat.MB, max: 5 * sizeFormat.GB}
	largeFileSizeConstraint = tempFileSizeConstraint{min: 10 * sizeFormat.GB, max: 60 * sizeFormat.GB}
)

//GenerateCmd starts the fs population process and recording of such process, if indicated by recorder
func GenerateCmd(ctx context.Context, rootPath string, size int64, recorder *IFileRecorder, errorChan chan<- error,
	writeFn writeFunc) *sync.WaitGroup {
	chanBuff := GetIntOrDefault(ctx, "max_parallel", 1)
	fmt.Println("generating using", chanBuff, "concurrent writers")

	workQueue := generateVolume(ctx, chanBuff, rootPath, size, errorChan)

	doneQueue := make(chan (*TempFile))
	var wg sync.WaitGroup
	wg.Add(chanBuff)
	genDoneCh := make(chan interface{})
	go func() {
		defer close(doneQueue)
		defer close(genDoneCh)

		if writeFn == nil {
			writeFn = writeVolume
		}
		//start file producing routines
		for i := 0; i < chanBuff; i++ {
			go writeFn(ctx, workQueue, doneQueue, &wg, errorChan)
		}

		wg.Wait()
	}()

	if recorder != nil {
		go recordVolume(ctx, recorder, doneQueue, errorChan)
		go reportGenerationProgressEveryMinute(ctx, size, recorder, genDoneCh)
	} else {
		go logProgressToStdout(ctx, doneQueue, size)
	}

	return &wg
}

func reportGenerationProgressEveryMinute(ctx context.Context, totalSize int64, recorder *IFileRecorder, exit chan interface{}) {
mainLoop:
	for {
		select {
		case <-ctx.Done():
			break mainLoop
		case <-exit:
			break mainLoop
		case <-time.After(1 * time.Minute):
			generated, err := (*recorder).GetTotalUnmarked()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				fmt.Printf("Generation: %2.3f%% done.\n", float64(generated*100)/float64(totalSize))
			}
		}
	}
}

func logProgressToStdout(ctx context.Context, doneQueue <-chan (*TempFile), totalSize int64) {
	processed := int64(0)
	for workItem := range processOrDone(ctx, doneQueue) {
		processed += workItem.size
		fmt.Printf("generated: %v. %2.3f%% done.\n", workItem, float64(processed*100)/float64(totalSize))
	}
}

func writeVolume(ctx context.Context, workQueue <-chan (*TempFile), doneQueue chan<- (*TempFile), wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	for workItem := range processOrDone(ctx, workQueue) {
		err := writeRandomFile(ctx, workItem)
		if err != nil {
			errChan <- err
		}
		doneQueue <- workItem
	}
}

func writeRandomFile(ctx context.Context, workItem *TempFile) error {
	fmt.Fprintln(os.Stdout, "generating", sizeFormat.ToString(workItem.size), workItem.path)
	fileHash, err := GenerateLen(ctx, workItem.size, workItem.path)
	if err != nil {
		return fmt.Errorf("error while generating %s: %v", workItem.path, err)
	}
	workItem.hash = fileHash

	return nil
}

//generateVolume generates the volume of TempFiles into channel it returns. async
func generateVolume(ctx context.Context, chanBuff int, basePath string, maxVolumeSize int64, errChan chan<- error) <-chan (*TempFile) {
	rand.Seed(time.Now().UnixNano())

	var maxTotalLargeFileSize int64 = int64(float64(maxVolumeSize) * maxLargeFilesShare)
	var maxTotalMedFileSize int64 = int64(float64(maxVolumeSize) * maxMedFilesShare)
	maxTotalSmallFileSize := maxVolumeSize - (maxTotalLargeFileSize + maxTotalMedFileSize)

	sizeGenerators := []func() int64{
		getRandomFileSizeFunc(&largeFileSizeConstraint, maxTotalLargeFileSize),
		getRandomFileSizeFunc(&medFileSizeConstraint, maxTotalMedFileSize),
		getRandomFileSizeFunc(&smallFileSizeConstraint, maxTotalSmallFileSize),
	}

	q := NewQueue()
	q.QueueEnqueue(volumePathFolder{
		basePath: basePath,
		filesNum: numFilesPerFolder,
	})

	workChan := make(chan (*TempFile), chanBuff)
	go func() {
		defer close(workChan)

		for q.QueueSize() != 0 && maxVolumeSize > 0 {
			select {
			case <-ctx.Done():
				fmt.Println("generateVolume: context exit")
			default:
			}
			queueElement, err := q.QueueDequeue()

			if err != nil {
				errChan <- err
				continue
			}

			path := (*queueElement).(volumePathFolder)
			maxVolumeSize = generateFilesForPathElement(&path, sizeGenerators, maxVolumeSize, workChan)

			if maxVolumeSize <= 0 {
				break
			}

			// more to generate in subfolders
			for i := 0; i < numSubfolders; i++ {
				_, err := q.QueueEnqueue(volumePathFolder{
					basePath: filepath.Join(path.basePath, fmt.Sprintf("subfolder_%d.tmp", i)),
					filesNum: numFilesPerFolder,
				})
				if err != nil {
					errChan <- err
					break
				}
			}
		}
	}()

	return workChan
}

func generateFilesForPathElement(
	pathElement *volumePathFolder,
	sizeGenerators []func() int64,
	maxVolumeSize int64,
	producerQueue chan<- (*TempFile),
) int64 {
	for pathElement.filesNum > 0 && maxVolumeSize > 0 {
		fileNumBeforeGenerators := pathElement.filesNum
		for _, sizeGen := range sizeGenerators {
			// if generated enough for this folder: back out
			if pathElement.filesNum == 0 {
				break
			}

			generatedFileSize := sizeGen()

			// if capped otherwise, or not enough space left, give opportunity for other generators to kick in
			if generatedFileSize == 0 || generatedFileSize > maxVolumeSize {
				continue
			}

			addToProducerQueue(generatedFileSize, pathElement, producerQueue)
			maxVolumeSize -= generatedFileSize

		}

		// corner case for last file in the volume, that might be too small.
		if pathElement.filesNum == fileNumBeforeGenerators && maxVolumeSize > 0 {
			addToProducerQueue(maxVolumeSize, pathElement, producerQueue)
			maxVolumeSize = 0
		}
	}

	return maxVolumeSize
}

func addToProducerQueue(size int64, pathElement *volumePathFolder, queue chan<- (*TempFile)) {
	pathToGenerateAt := filepath.Join(pathElement.basePath, fmt.Sprintf("file_%d.tmp", pathElement.filesNum))
	queue <- &TempFile{path: pathToGenerateAt, size: size}
	pathElement.filesNum--
}

func getRandomFileSizeFunc(minMaxConstraint *tempFileSizeConstraint, capConstraint int64) func() int64 {
	return func() int64 {
		if minMaxConstraint.min > capConstraint {
			return 0
		}

		retVal := minMaxConstraint.min + rand.Int63n(minMaxConstraint.max-minMaxConstraint.min)

		if retVal > capConstraint {
			return capConstraint
		}

		return retVal
	}
}

func pseudoWriteFile(ctx context.Context, workQueue <-chan (*TempFile), doneQueue chan<- (*TempFile), wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	hashLen := 10
	for workItem := range processOrDone(ctx, workQueue) {
		lenPath := len(workItem.path)
		if lenPath >= hashLen {
			workItem.hash = workItem.path[lenPath-hashLen:]
		} else {
			workItem.hash = workItem.path
		}

		doneQueue <- workItem
	}
}
