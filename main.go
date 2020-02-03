package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/karrick/godirwalk"
)

var verbose = false
var debug = false
var countSuccess int64 = 0
var countSkippedDirs int64 = 0

func traverse(dir string) error {
	err := godirwalk.Walk(dir, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			countSuccess++

			os.Lstat(osPathname)

			if debug {
				// ftype := "f"
				// if de.ModeType().IsDir() {
				// 	ftype = "d"
				// }

				// symlink := ""
				// if (de.ModeType() & os.ModeSymlink) != 0 {
				// 	symlink = "(s)"
				// }

				// fmt.Printf("%v%v\t%v/%v\t(%v bytes)\tmod %v\n",
				// 	ftype, symlink, dir, de.Name(), fileinfo.Size(), fileinfo.ModTime())
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

			return nil
		},
		ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
			if verbose {
				fmt.Printf("\nFailed to traverse '%v': %v\n", path, err)
			}
			countSkippedDirs++
			return godirwalk.SkipNode
		},
		Unsorted:            true,
		FollowSymbolicLinks: false,
	})
	return err
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
