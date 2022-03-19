package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

// File is the path to the script
var File string

// Log files for optional performance tracing.
var CPUProfile, MemProfile, Trace string

func main() {
	progname := os.Args[0]
	log.SetFlags(0)
	log.SetPrefix(progname + ": ")

	RegisterFlags()

	flag.Parse()

	args := flag.Args()

	log.Println("**==** ", File)

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, `%[1]s is a tool for transforming text.

Usage: %[1]s [-file] [/path/to/script.yaml]

Run '%[1]s help' for more detail,
`, progname)
	}
	sc := CLI(args)
	os.Exit(sc)
}

func RegisterFlags() {
	flag.StringVar(&File, "file", "f", "file of script (defaults to carbon.yaml)")

	flag.StringVar(&CPUProfile, "cpuprofile", "", "write CPU profile to this file")
	flag.StringVar(&MemProfile, "memprofile", "", "write memory profile to this file")
	flag.StringVar(&Trace, "trace", "", "write trace log to this file")
}

func CLI(args []string) (exitcode int) {
	if CPUProfile != "" {
		f, err := os.Create(CPUProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		// profile won't be written in case of error.
		defer pprof.StopCPUProfile()
	}
	if MemProfile != "" {
		f, err := os.Create(MemProfile)
		if err != nil {
			log.Fatal(err)
		}
		// NB: memprofile won't be written in case of error.
		defer func() {
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatalf("Writing memory profile: %v", err)
			}
			f.Close()
		}()
	}
	if Trace != "" {
		f, err := os.Create(Trace)
		if err != nil {
			log.Fatal(err)
		}
		if err := trace.Start(f); err != nil {
			log.Fatal(err)
		}
		// NB: trace log won't be written in case of error.
		defer func() {
			trace.Stop()
			log.Printf("To view the trace, run:\n$ go tool trace view %s", Trace)
		}()
	}

	run()

	return 0
}
