package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	sizeFormat "github.com/rdev02/size-format"
)

const (
	verifyInMem    = "mem"
	verifyInSQLite = "sqlite"
)

type (
	cmdFlags struct {
		rootPath string
		size     uint64
		verify   string
	}

	//TempFile connects main/generator/processor and recorder
	TempFile struct {
		path string
		size uint64
		hash string
	}
)

func defaultFlags() *cmdFlags {
	defaults := cmdFlags{
		rootPath: ".",
		size:     sizeFormat.GB,
		verify:   verifyInMem,
	}

	return &defaults
}

func main() {
	cmdFlags := defaultFlags()
	flag.Uint64Var(&cmdFlags.size, "size", cmdFlags.size, "the size to test with. set this to the volume size. default 1GB")
	flag.StringVar(&cmdFlags.verify, "verify", cmdFlags.verify, "verify results via")

	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Println("path not provided. syntax: disktest [opts] path")
		flag.PrintDefaults()
	}

	chanBuff := runtime.NumCPU() - 1
	if chanBuff == 0 {
		chanBuff = 1
	}
	workQueue := make(chan (*TempFile), chanBuff)
	doneQueue := make(chan (*TempFile))
	defer close(workQueue)

	ctx := context.Background()

	// start files generation routine
	go GenerateVolume(ctx, workQueue, flag.Args()[0], cmdFlags.size)

	//start file producing routines
	for i := 0; i < chanBuff; i++ {
		go writeVolume(ctx, workQueue)
	}

	// start recording routine
	go recordVolume(ctx, doneQueue)
}

func recordVolume(ctx context.Context, doneQueue <-chan (*TempFile)) {
	for {
		select {
		case <-ctx.Done():
			break
		case generatedFile, ok := <-doneQueue:
			if !ok {
				break
			}

			recordRandomFile(generatedFile)
		}
	}
}

func writeVolume(ctx context.Context, workQueue <-chan (*TempFile)) <-chan (*TempFile) {
	for {
		select {
		case <-ctx.Done():
			break
		case workItem, ok := <-workQueue:
			if !ok {
				break
			}
			writeRandomFile(ctx, workItem)
		}
	}
}

func writeRandomFile(ctx context.Context, workItem *TempFile) {
	fmt.Fprintln(os.Stdout, "generating", sizeFormat.ToString(workItem.size), "at", workItem.path)
	fileHash, err := GenerateLen(workItem.size, workItem.path)
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Errorf("error while generating %s: %v", workItem.path, err))
		fmt.Println("cancelling execution")
		_, cancel := context.WithCancel(ctx)
		cancel()
	}
	workItem.hash = fileHash
}

func recordRandomFile(tmpFile *TempFile) {
	fmt.Fprintln(os.Stdout, "Storing result for", tmpFile.path, "size", tmpFile.size, "hash", tmpFile.hash)
}
