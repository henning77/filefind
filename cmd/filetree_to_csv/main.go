package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/henning77/filefind/internal/util"
)

var verbose = false
var debug = false
var countSuccess int64 = 0
var countSkippedDirs int64 = 0

func traverse(dir string, basedir string, csvWriter *csv.Writer) error {
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
		// Skip non-files/-dirs/-symlinks
		if fileinfo.Mode()&(os.ModeNamedPipe|os.ModeSocket|os.ModeDevice|os.ModeCharDevice|os.ModeIrregular) != 0 {
			if debug {
				fmt.Fprintf(os.Stderr, "\nSkipped non-file/-dir/-symlink: %v/%v\n", dir, fileinfo.Name())
			}
			continue
		}

		if util.FileToExclude(fileinfo.Name()) {
			if debug {
				fmt.Fprintf(os.Stderr, "\nExcluded: %v/%v\n", dir, fileinfo.Name())
			}
			continue
		}

		isDir := 0
		if fileinfo.IsDir() {
			isDir = 1

			subdir := path.Join(dir, fileinfo.Name())
			if err := traverse(subdir, basedir, csvWriter); err != nil {
				fmt.Fprintf(os.Stderr, "\nFailed to traverse '%v': %v\n", subdir, err)
				countSkippedDirs++
			}
		}

		isSymlink := 0
		if (fileinfo.Mode() & os.ModeSymlink) != 0 {
			isSymlink = 1
		}

		reldir := dir
		if basedir != "/" {
			reldir, err = filepath.Rel(basedir, dir)
			if err != nil {
				return err
			}
		}

		csvWriter.Write([]string{
			reldir,
			fileinfo.Name(),
			util.LowercaseExtension(fileinfo.Name()),
			strconv.Itoa(isDir),
			strconv.Itoa(isSymlink),
			strconv.FormatInt(fileinfo.Size(), 10),
			fileinfo.ModTime().Format("'2006-01-02 15:04:05'"),
		})

		countSuccess++

		if debug {
			fmt.Fprintf(os.Stderr, "%v%v %v\t%v/%v (%v)\tmod %v\n",
				isDir, isSymlink, fileinfo.Mode(), dir, fileinfo.Name(), fileinfo.Size(), fileinfo.ModTime())
		} else {
			if countSuccess%1000 == 0 {
				fmt.Fprint(os.Stderr, ".")
			}
			if countSuccess%100000 == 0 {
				fmt.Fprint(os.Stderr, "\n")
			}
			if countSuccess%1000000 == 0 {
				fmt.Fprint(os.Stderr, "\n")
			}
		}
	}

	return nil
}

func main() {
	pathPtr := flag.String("path", ".", "Path of the root directory to traverse")
	basepathPtr := flag.String("base", "/", "Base path to strip from all paths")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Debug output")
	flag.Parse()

	path, err := filepath.Abs(*pathPtr)
	if err != nil {
		log.Fatal(err)
	}

	basepath, err := filepath.Abs(*basepathPtr)
	if err != nil {
		log.Fatal(err)
	}

	csvWriter := csv.NewWriter(os.Stdout)

	start := time.Now()
	if err := traverse(path, basepath, csvWriter); err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)

	csvWriter.Flush()

	fps := float64(countSuccess) / elapsed.Seconds()
	fmt.Fprintf(os.Stderr, "\nFound %v files and dirs. Skipped %v directories. Time %.1fs (%.1f files per second)\n",
		countSuccess, countSkippedDirs, elapsed.Seconds(), fps)
}
