package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	sizeFormat "github.com/rdev02/size-format"
)

const (
	verifyInMem    = "mem"
	verifyInSQLite = "sqlite"
)

type (
	cmdFlags struct {
		rootPath string
		size     int64
		verify   string
		generate string
	}

	//TempFile connects main/generator/processor and recorder
	TempFile struct {
		path string
		size int64
		hash string
	}

	IFileRecorder interface {
		RecordFile(file *TempFile) error
		MarkFileExits(file *TempFile) (bool, error)
		VerifyFileExits(file *TempFile) (bool, error)
		FilesNotCheckedYet() ([]*TempFile, error)
	}
)

func (tf *TempFile) String() string {
	return fmt.Sprint(tf.path, " size: ", sizeFormat.ToString(uint64(tf.size)), " hash: ", tf.hash)
}

func defaultFlags() *cmdFlags {
	defaults := cmdFlags{
		rootPath: ".",
		size:     sizeFormat.GB,
		verify:   verifyInMem,
		generate: "y",
	}

	return &defaults
}

func main() {
	cmdFlags := defaultFlags()
	flag.Int64Var(&cmdFlags.size, "size", cmdFlags.size, "the total size of files to generate. no effect if used without the --generate flag")
	flag.StringVar(&cmdFlags.generate, "generate", cmdFlags.generate, "generate files at the location specified: y/n")
	flag.StringVar(&cmdFlags.verify, "verify", cmdFlags.verify, fmt.Sprintf("verify results via %s/%s/none", verifyInMem, verifyInSQLite))

	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Println("path not provided. syntax: disktest [opts] path")
		flag.PrintDefaults()
		return
	}

	var recordingStrategy IFileRecorder = nil
	switch cmdFlags.verify {
	case verifyInMem:
		recordingStrategy = NewInMemRecorder()
	case verifyInSQLite:
		recordingStrategy = NewSqlLiteRecorder()
	default:
	}

	ctx, stopExecution := context.WithCancel(context.Background())

	// start files generation routine
	rootPath := flag.Args()[0]
	if len(rootPath) == 0 {
		rootPath = cmdFlags.rootPath
	}

	var errorChan = make(chan error)

	var generateDone *sync.WaitGroup
	if strings.Compare(cmdFlags.generate, "y") == 0 {
		generateDone = GenerateCmd(ctx, cmdFlags.rootPath, cmdFlags.size, &recordingStrategy, errorChan)
	}

	var verifyDone *sync.WaitGroup
	if len(cmdFlags.verify) > 0 {
		go func() {
			// verify strictly after all recording has been done
			if generateDone != nil {
				generateDone.Wait()
			}

			verifyDone = VerifyCmd(ctx, &recordingStrategy, cmdFlags.rootPath, errorChan)
		}()
	}

	select {
	case <-ctx.Done():
		stopExecution()
		fmt.Fprintln(os.Stderr, ctx.Err())
	case numDone := <-waitForAllCommands(generateDone, verifyDone):
		fmt.Println(numDone, "tasks completed")
	}

	fmt.Println("All done, exiting.")
}

func waitForAllCommands(cmds ...*sync.WaitGroup) chan rune {
	res := make(chan rune)
	var cnt rune

	go func() {
		defer close(res)
		for _, cmd := range cmds {
			if cmd != nil {
				cmd.Wait()
				cnt++
			}
		}
		res <- cnt
	}()

	return res
}

func processOrDone(ctx context.Context, ch <-chan (*TempFile)) <-chan (*TempFile) {
	res := make(chan (*TempFile))

	go func() {
		defer close(res)
	main:
		for {
			select {
			case <-ctx.Done():
				fmt.Fprintln(os.Stderr, "context interrupt")
				fmt.Fprintln(os.Stderr, ctx.Err())
				break main
			case workItem, ok := <-ch:
				if !ok {
					break main
				}
				select {
				case res <- workItem:
				case <-ctx.Done():
				}
			}
		}
	}()

	return res
}
