package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

var verbose = false
var debug = false
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

func escapeSql(str string) string {
	return strings.ReplaceAll(str, "'", "''")
}

func traverseAndInsert(dir string) error {
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
				fmt.Printf("\nSkipped non-file/-dir/-symlink: %v/%v\n", dir, fileinfo.Name())
			}
			continue
		}

		isDir := 0
		if fileinfo.IsDir() {
			isDir = 1

			subdir := path.Join(dir, fileinfo.Name())
			if err := traverseAndInsert(subdir); err != nil {
				if verbose || debug {
					fmt.Printf("\nFailed to traverse '%v': %v\n", subdir, err)
				}
				countSkippedDirs++
			}
		}

		isSymlink := 0
		if (fileinfo.Mode() & os.ModeSymlink) != 0 {
			isSymlink = 1
		}

		fmt.Printf("INSERT INTO file(base, name, ext, is_dir, is_symlink, size, modified) VALUES('%s', '%s', '%s', %d, %d, %d, %s);\n",
			escapeSql(dir),
			escapeSql(fileinfo.Name()),
			escapeSql(lowercaseExtension(fileinfo.Name())),
			isDir,
			isSymlink,
			fileinfo.Size(),
			fileinfo.ModTime().Format("'2006-01-02 15:04:05'"),
		)

		countSuccess++

		if debug {
			fmt.Printf("%v%v %v\t%v/%v (%v)\tmod %v\n",
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

func createDbSchema() error {
	sqlStmt := `
		CREATE TABLE file (
			base TEXT NOT NULL, 
			name TEXT NOT NULL,
			ext  TEXT NOT NULL, -- always lowercase
			is_dir     INTEGER NOT NULL, -- 0: false, 1: true
			is_symlink INTEGER NOT NULL, -- 0: false, 1: true
			size       INTEGER NOT NULL,
			modified   INTEGER NOT NULL -- Unix time (seconds)
		);
	`

	fmt.Println(sqlStmt)
	return nil
}

func main() {
	pathPtr := flag.String("path", ".", "Path of the root directory to traverse")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Debug output")
	flag.Parse()

	if err := createDbSchema(); err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	if err := traverseAndInsert(*pathPtr); err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)

	fps := float64(countSuccess) / elapsed.Seconds()
	fmt.Fprintf(os.Stderr, "\nFound %v files and dirs. Skipped %v directories. Time %.1fs (%.1f files per second)\n",
		countSuccess, countSkippedDirs, elapsed.Seconds(), fps)
}
