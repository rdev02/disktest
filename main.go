package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
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
		rootPath   string
		size       int64
		verify     string
		generate   string
		cpuprofile string
		memprofile string
	}

	//TempFile connects main/generator/processor and recorder
	TempFile struct {
		path string
		size int64
		hash string
	}

	//IFileRecorder defines methods necessary to record a file
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
	flag.StringVar(&cmdFlags.cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&cmdFlags.memprofile, "memprofile", "", "write mem profile to file")

	flag.Parse()

	// cpu profiling
	if cmdFlags.cpuprofile != "" {
		f, err := os.Create(cmdFlags.cpuprofile)
		if err != nil {
			fmt.Println(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		defer f.Close()
	}

	// mem profiling
	if cmdFlags.memprofile != "" {
		f, err := os.Create(cmdFlags.memprofile)
		if err != nil {
			fmt.Println(err)
		}
		defer pprof.WriteHeapProfile(f)
		defer f.Close()
	}

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
	defer close(errorChan)

	var generateDone *sync.WaitGroup
	if strings.Compare(cmdFlags.generate, "y") == 0 {
		generateDone = GenerateCmd(ctx, cmdFlags.rootPath, cmdFlags.size, &recordingStrategy, errorChan, nil)
	}

	var verifyDone *sync.WaitGroup
	if len(cmdFlags.verify) > 0 {
		go func() {
			// verify strictly after all recording has been done
			if generateDone != nil {
				generateDone.Wait()
			}

			wg, err := VerifyCmd(ctx, &recordingStrategy, cmdFlags.rootPath, errorChan)
			if err != nil {
				panic(err)
			}

			verifyDone = wg
		}()
	}

loop:
	for {
		select {
		case err, ok := <-errorChan:
			//for now die on any error
			if err != nil || !ok {
				fmt.Fprintln(os.Stderr, err)
				stopExecution()
			}
			break loop
		case <-ctx.Done():
			stopExecution()
			fmt.Fprintln(os.Stderr, ctx.Err())
			break loop
		case numDone := <-waitForAllCommands(generateDone, verifyDone):
			fmt.Println(numDone, "tasks completed")
			break loop
		}
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
