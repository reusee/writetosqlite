package main

import "database/sql"

func initDB(db *sql.DB) error {

	_, err := db.Exec(`
  create table if not exists files (
    id integer auto increment,
    path text unique,
    mode unsigned big int,
    size big int
  )
  `)
	ce(err)

	_, err = db.Exec(`
  create table if not exists blobs (
    id integer auto increment,
    content blob,
    sha256 text
  )
  `)
	ce(err)

	return nil
}
