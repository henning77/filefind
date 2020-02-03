package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"
)

var verbose = false
var debug = false
var countSuccess int64 = 0
var countSkippedDirs int64 = 0

func traverse(dir string) error {

	root, err := os.Open(dir)
	if err != nil {
		return err
	}

	fileinfos, err := root.Readdir(0)
	if err != nil {
		return err
	}

	if err := root.Close(); err != nil {
		return err
	}

	for _, fileinfo := range fileinfos {
		if fileinfo.IsDir() {
			subdir := path.Join(dir, fileinfo.Name())

			if err := traverse(subdir); err != nil {
				if verbose {
					fmt.Printf("\nFailed to traverse '%v': %v\n", subdir, err)
				}
				countSkippedDirs++
			}
		}

		countSuccess++

		if debug {
			ftype := "f"
			if fileinfo.IsDir() {
				ftype = "d"
			}

			symlink := ""
			if (fileinfo.Mode() & os.ModeSymlink) != 0 {
				symlink = "(s)"
			}

			fmt.Printf("%v%v\t%v/%v\t(%v bytes)\tmod %v\n",
				ftype, symlink, dir, fileinfo.Name(), fileinfo.Size(), fileinfo.ModTime())
		} else {
			if countSuccess%1000 == 0 {
				fmt.Print(".")
			}
			if countSuccess%100000 == 0 {
				fmt.Print("\n")
			}
			if countSuccess%1000000 == 0 {
				fmt.Print("\n")
			}
		}
	}

	return nil
}

func main() {
	pathPtr := flag.String("path", ".", "Path of the root directory to traverse")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Debug output")
	flag.Parse()

	start := time.Now()
	err := traverse(*pathPtr)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("\nFailed: %v\n", err)
		os.Exit(1)
	}

	fps := float64(countSuccess) / elapsed.Seconds()
	fmt.Printf("\nFound %v files and dirs. Skipped %v directories. Time %.1fs (%.1f files per second)\n",
		countSuccess, countSkippedDirs, elapsed.Seconds(), fps)
}
