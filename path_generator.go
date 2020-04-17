package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	sizeFormat "github.com/rdev02/size-format"
)

type (
	volumePathFolder struct {
		basePath string
		filesNum rune
	}

	tempFileSizeConstraint struct {
		min, max uint64
	}
)

const (
	numFilesPerFolder  = 1001
	numSubfolders      = 10
	maxLargeFilesShare = .5
	maxMedFilesShare   = .35
)

var (
	smallFileSizeConstraint = tempFileSizeConstraint{min: 100 * sizeFormat.KB, max: 50 * sizeFormat.MB}
	medFileSizeConstraint   = tempFileSizeConstraint{min: 100 * sizeFormat.MB, max: 5 * sizeFormat.GB}
	largeFileSizeConstraint = tempFileSizeConstraint{min: 10 * sizeFormat.GB, max: 60 * sizeFormat.GB}
)

//GenerateVolume generates volume of maxSize files in basePath
func GenerateVolume(ctx context.Context, workChan chan<- (*TempFile), basePath string, maxVolumeSize uint64) {
	rand.Seed(time.Now().UnixNano())

	var maxTotalLargeFileSize uint64 = uint64(float64(maxVolumeSize) * maxLargeFilesShare)
	var maxTotalMedFileSize uint64 = uint64(float64(maxVolumeSize) * maxMedFilesShare)
	maxTotalSmallFileSize := maxVolumeSize - (maxTotalLargeFileSize + maxTotalMedFileSize)

	sizeGenerators := []func() uint64{
		getRandomFileSizeFunc(&largeFileSizeConstraint, maxTotalLargeFileSize),
		getRandomFileSizeFunc(&medFileSizeConstraint, maxTotalMedFileSize),
		getRandomFileSizeFunc(&smallFileSizeConstraint, maxTotalSmallFileSize),
	}

	q := NewQueue()
	q.QueueEnqueue(volumePathFolder{
		basePath: basePath,
		filesNum: numFilesPerFolder,
	})

	for q.size != 0 && maxVolumeSize > 0 {
		select {
		case <-ctx.Done():
			fmt.Println("generateVolume: context exit")
		default:
		}
		queueElement, err := q.QueueDequeue()

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "GenerateVolume: cancelling execution")
			_, cancel := context.WithCancel(ctx)
			cancel()
			break
		}

		path := (*queueElement).(volumePathFolder)
		maxVolumeSize = generateFilesForPathElement(ctx, &path, sizeGenerators, maxVolumeSize, workChan)

		if maxVolumeSize <= 0 {
			break
		}

		// more to generate in subfolders
		for i := 0; i < numSubfolders; i++ {
			q.QueueEnqueue(volumePathFolder{
				basePath: filepath.Join(path.basePath, fmt.Sprintf("subfolder_%d.tmp", i)),
				filesNum: numFilesPerFolder,
			})
		}
	}
}

func generateFilesForPathElement(
	ctx context.Context,
	pathElement *volumePathFolder,
	sizeGenerators []func() uint64,
	maxVolumeSize uint64,
	producerQueue chan<- (*TempFile),
) uint64 {
	for pathElement.filesNum > 0 && maxVolumeSize > 0 {
		fileNumBeforeGenerators := pathElement.filesNum
		for _, sizeGen := range sizeGenerators {
			// if generated enough for this folder: back out
			if pathElement.filesNum == 0 {
				break
			}

			select {
			case <-ctx.Done():
				return maxVolumeSize
			default:
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

func addToProducerQueue(size uint64, pathElement *volumePathFolder, queue chan<- (*TempFile)) {
	pathToGenerateAt := filepath.Join(pathElement.basePath, fmt.Sprintf("file_%d.tmp", pathElement.filesNum))
	queue <- &TempFile{path: pathToGenerateAt, size: size}
	pathElement.filesNum--
}

func getRandomFileSizeFunc(minMaxConstraint *tempFileSizeConstraint, capConstraint uint64) func() uint64 {
	return func() uint64 {
		if minMaxConstraint.max > capConstraint {
			return 0
		}

		return minMaxConstraint.min + uint64(float64(minMaxConstraint.max-minMaxConstraint.max)*rand.Float64())
	}
}
