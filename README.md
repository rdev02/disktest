# disktest
Small go utility to fill FS with randomly generated files, the read them back and verify MD5 hash matches. Uses very modest amount of RAM, especially when run inside the container.

## Usage
```
$ go build
$ ./disktest
path not provided. syntax: disktest [opts] path
  -cpuprofile string
    	write cpu profile to file
  -generate string
    	generate files at the location specified: y/n (default "y")
  -maxparallel int
    	max parallel processing streams. default(0) = CPU cores - 1
  -memprofile string
    	write mem profile to file
  -size string
    	the total size of files to generate. no effect if used without the --generate flag (default "1GB")
  -verify string
    	verify results via mem/sqlite/none (default "mem")
  -waitbeforeexit string
    	wait before exiting y/n (default "n")
```

## examples

`./disktest -size=5.5GB .`
would generate 5.5 GB worth of random files, using CPU-1 concurrent threads. Then verify them and print results.

`./disktest -size=0.5TB -verify=n -maxparallel=1 /var/temp/`
will generate 0.5 TB worth of random files without verification in `/var/temp`

## docker
Provided `Dockerfile` assumes you have prebuilt disktest binary with `go build`. For Alpine you can do this with `docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:alpine go build -v`. See the docker file for ENV variable overrides.