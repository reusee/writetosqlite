package main

import "os"

func main() {
	sqlitePath := os.Args[1]
	filePaths := os.Args[2:]
	ce(write(sqlitePath, filePaths))
}
