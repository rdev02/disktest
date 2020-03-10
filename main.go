package main

import (
	"flag"
	"fmt"

	sizeFormat "github.com/rdev02/size-format"
)

const (
	verifyInMem    = "mem"
	verifyInSQLite = "sqlite"
)

type cmdFlags struct {
	rootPath string
	size     uint64
	verify   string
}

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

	res, err := generateLen(1.5*sizeFormat.KB, "./tst")
	if err != nil {
		fmt.Errorf("could not write %v", err)
		return
	}

	fmt.Printf("MD5: %x \n", res)

	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Println("path not provided. syntax: disktest [opts] path")
		flag.PrintDefaults()
	}
}
