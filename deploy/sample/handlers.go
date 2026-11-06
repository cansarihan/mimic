// Package sample is a small, deliberately flawed codebase used to demonstrate
// Mimic. Each function carries one of the shapes the rules look for, and the two
// validate* functions are near-duplicates so `mimic like` has a sibling to find.
package sample

import (
	"database/sql"
	"fmt"
	"os"
)

// readAll leaks file handles: the defer runs when readAll returns, not per loop.
func readAll(paths []string) int {
	total := 0
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()
		info, _ := f.Stat()
		total += int(info.Size())
	}
	return total
}

// lookup builds its query with Sprintf — an injection waiting to happen.
func lookup(db *sql.DB, name string) (*sql.Rows, error) {
	return db.Query(fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name))
}

// save swallows the error it just checked.
func save(db *sql.DB, name string) {
	_, err := db.Exec("INSERT INTO users(name) VALUES($1)", name)
	if err != nil {
	}
}
