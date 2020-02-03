package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
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

func traverseAndInsert(dir string, insertStatement *sql.Stmt) error {
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
			if err := traverseAndInsert(subdir, insertStatement); err != nil {
				if verbose {
					fmt.Printf("\nFailed to traverse '%v': %v\n", subdir, err)
				}
				countSkippedDirs++
			}
		}

		isSymlink := 0
		if (fileinfo.Mode() & os.ModeSymlink) != 0 {
			isSymlink = 1
		}

		_, err = insertStatement.Exec(
			dir,
			fileinfo.Name(),
			lowercaseExtension(fileinfo.Name()),
			isDir,
			isSymlink,
			fileinfo.Size(),
			fileinfo.ModTime(),
		)
		if err != nil {
			log.Fatal(err)
		}

		countSuccess++

		if debug {
			fmt.Printf("%v%v\t%v/%v\t(%v bytes)\tmod %v\n",
				isDir, isSymlink, dir, fileinfo.Name(), fileinfo.Size(), fileinfo.ModTime())
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

func createDbSchema(db *sql.DB) error {
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
	if _, err := db.Exec(sqlStmt); err != nil {
		return err
	}

	return nil
}

func main() {
	pathPtr := flag.String("path", ".", "Path of the root directory to traverse")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&debug, "debug", false, "Debug output")
	flag.Parse()

	os.Remove("./filefind.db")
	db, err := sql.Open("sqlite3", "./filefind.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := createDbSchema(db); err != nil {
		log.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO file(base, name, ext, is_dir, is_symlink, size, modified) VALUES(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	start := time.Now()
	if err := traverseAndInsert(*pathPtr, stmt); err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)

	tx.Commit()

	fps := float64(countSuccess) / elapsed.Seconds()
	fmt.Printf("\nFound %v files and dirs. Skipped %v directories. Time %.1fs (%.1f files per second)\n",
		countSuccess, countSkippedDirs, elapsed.Seconds(), fps)
}
