package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func write(
	sqliteFile string,
	filePaths []string,
) error {

	sqliteFile, err := filepath.Abs(sqliteFile)
	ce(err)
	for i, path := range filePaths {
		path, err := filepath.Abs(path)
		ce(err)
		filePaths[i] = path
	}

	db, err := sql.Open("sqlite3", sqliteFile)
	ce(err)
	defer db.Close()
	ce(initDB(db))

	for _, filePath := range filePaths {
		ce(filepath.WalkDir(filePath, func(fullPath string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if entry.IsDir() {
				return nil
			}

			if fullPath == sqliteFile {
				return nil
			}
			info, err := entry.Info()
			ce(err)

			tx, err := db.Begin()
			ce(err)
			var existedSize int64
			err = tx.QueryRow(`
        select size
        from files
        where path = ?
        `,
				fullPath,
			).Scan(&existedSize)
			if err == nil {
				// existed
				if existedSize != info.Size() {
					ce(fmt.Errorf("size not match: %v", fullPath))
				}
				tx.Rollback()
				return nil
			}

			pt("write %v\n", fullPath)

			if !errors.Is(err, sql.ErrNoRows) {
				ce(err)
			}

			content, err := os.ReadFile(fullPath)
			ce(err)
			sha256sum := fmt.Sprintf("%x", sha256.Sum256(content))

			var n int
			err = tx.QueryRow(`
        select count(*) from blobs
        where sha256 = ?
        `,
				sha256sum,
			).Scan(&n)
			ce(err)
			if n == 0 {
				// write blob
				_, err = tx.Exec(`
          insert into blobs (
            content, sha256
          ) values (
            ?, ?
          )
          `,
					content,
					sha256sum,
				)
				ce(err)
			}

			if len(content) > 1_000_000_000 {
				ce(fmt.Errorf("file too large: %v", filePath))
			}

			_, err = tx.Exec(`
        insert into files (
          path, mode, size
        ) values (
          ?, ?, ?
        )
        `,
				fullPath,
				info.Mode(),
				info.Size(),
			)
			ce(err)

			ce(tx.Commit())

			return nil
		}))
	}

	return nil
}
