package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var verbose = false
var debug = false
var maxSizeToCopy int64 = 50 * 1024
var fileExtensionsToCopy = []string{"txt", "md", "bas", "c", "cpp", "h", "java", "go", "doc", "docx", "xls", "xlsx"}

var countSuccess int64 = 0
var countSkippedDirs int64 = 0

func lowercaseExtension(filename string) string {
	ext := path.Ext(filename)
	// Remove dot
	if len(ext) > 1 {
		ext = ext[1:]
	}
	return strings.ToLower(ext)
}

func fileExtensionEligibleForCopy(filename string) bool {
	ext := lowercaseExtension(filename)
	for _, e := range fileExtensionsToCopy {
		if e == ext {
			return true
		}
	}
	return false
}

func fileEligibleForCopy(file os.FileInfo) bool {
	return fileExtensionEligibleForCopy(file.Name()) && file.Size() <= maxSizeToCopy
}

func createFileProxy(srcAbsDir string, srcBaseDir string, destAbsDir string, source os.FileInfo) error {
	destRelDir, err := filepath.Rel(srcBaseDir, srcAbsDir)
	if err != nil {
		return err
	}

	destFullDir := filepath.Join(destAbsDir, destRelDir)
	if err := os.MkdirAll(destFullDir, 0770); err != nil {
		return err
	}

	destFilename := filepath.Join(destFullDir, source.Name())

	dest, err := os.Create(destFilename)
	if err != nil {
		return err
	}
	defer dest.Close()

	if fileEligibleForCopy(source) {
		source, err := os.Open(path.Join(srcAbsDir, source.Name()))
		if err != nil {
			return err
		}
		defer source.Close()

		_, err = io.Copy(dest, source)
		if err != nil {
			return err
		}
	}

	if err := os.Chtimes(destFilename, source.ModTime(), source.ModTime()); err != nil {
		return err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Created %v\n", destFilename)
	}

	return nil
}

func traverse(dir string, baseDir string, destDir string) error {
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

		isDir := 0
		if fileinfo.IsDir() {
			isDir = 1

			subdir := path.Join(dir, fileinfo.Name())
			if err := traverse(subdir, baseDir, destDir); err != nil {
				if verbose || debug {
					fmt.Fprintf(os.Stderr, "\nFailed to traverse '%v': %v\n", subdir, err)
				}
				countSkippedDirs++
			}
		} else {
			if err := createFileProxy(dir, baseDir, destDir, fileinfo); err != nil {
				return err
			}
		}

		countSuccess++

		if debug {
			fmt.Fprintf(os.Stderr, "%v %v\t%v/%v (%v)\tmod %v\n",
				isDir, fileinfo.Mode(), dir, fileinfo.Name(), fileinfo.Size(), fileinfo.ModTime())
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
	srcDirPtr := flag.String("src", ".", "Source directory")
	destDirPtr := flag.String("dest", "", "Destination directory")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Debug output")
	flag.Parse()

	start := time.Now()
	if err := traverse(*srcDirPtr, *srcDirPtr, *destDirPtr); err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)

	fps := float64(countSuccess) / elapsed.Seconds()
	fmt.Fprintf(os.Stderr, "\nFound %v files and dirs. Skipped %v directories. Time %.1fs (%.1f files per second)\n",
		countSuccess, countSkippedDirs, elapsed.Seconds(), fps)
}
