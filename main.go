package main

import (
	"bufio"
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
		rootPath       string
		size           string
		verify         string
		generate       string
		cpuprofile     string
		memprofile     string
		waitbeforeexit string
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
		rootPath:       ".",
		size:           "1GB",
		verify:         verifyInMem,
		generate:       "y",
		waitbeforeexit: "n",
	}

	return &defaults
}

func main() {
	cmdFlags := defaultFlags()
	flag.StringVar(&cmdFlags.size, "size", cmdFlags.size, "the total size of files to generate. no effect if used without the --generate flag")
	flag.StringVar(&cmdFlags.generate, "generate", cmdFlags.generate, "generate files at the location specified: y/n")
	flag.StringVar(&cmdFlags.verify, "verify", cmdFlags.verify, fmt.Sprintf("verify results via %s/%s/none", verifyInMem, verifyInSQLite))
	flag.StringVar(&cmdFlags.cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&cmdFlags.memprofile, "memprofile", "", "write mem profile to file")
	flag.StringVar(&cmdFlags.waitbeforeexit, "waitbeforeexit", cmdFlags.waitbeforeexit, "wait before exiting y/n")

	flag.Parse()

	sizeBytes, err := sizeFormat.ToNum(&cmdFlags.size)
	if err != nil {
		panic(err)
	}
	fmt.Println("will generate", sizeFormat.ToString(sizeBytes))

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

	var recordingStrategy *IFileRecorder
	switch cmdFlags.verify {
	case verifyInMem:
		rec := IFileRecorder(NewInMemRecorder())
		recordingStrategy = &rec
		fmt.Println("using in-memory recorder")
	case verifyInSQLite:
		rec := IFileRecorder(NewSqlLiteRecorder())
		recordingStrategy = &rec
		fmt.Println("using SqLite recorder")
	default:
		fmt.Println("no recording")
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
		fmt.Println("preparing to generate files")
		generateDone = GenerateCmd(ctx, rootPath, int64(sizeBytes), recordingStrategy, errorChan, nil)
	}

	var verifyDone *sync.WaitGroup
	if len(cmdFlags.verify) > 0 && recordingStrategy != nil {
		fmt.Println("preparing to verify files")

		// verify strictly after all recording has been done
		if generateDone != nil {
			generateDone.Wait()
		}

		wg, err := VerifyCmd(ctx, recordingStrategy, rootPath, errorChan)
		if err != nil {
			panic(err)
		}

		verifyDone = wg
	} else {
		fmt.Println("no verification. please check your -verify flag")
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

	fmt.Println("All done, exiting")
	if strings.Compare(cmdFlags.waitbeforeexit, "y") == 0 {
		fmt.Println("Press return to exit...")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadLine()
	}
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
