package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

const PROGRAMVERSION = "Rupert v0.1 Copyright © 2025 Bob Stammers"

// DBNAME names the database file
var DBNAME *string = flag.String("db", "ibaukrd.db", "database file")

// HTTPPort is the web port to serve
var HTTPPort *string = flag.String("port", "1080", "Web port")

// DBH provides access to the database
var DBH *sql.DB

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func getIntegerFromDB(sqlx string, defval int64) int64 {

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	if rows.Next() {
		var val int64
		rows.Scan(&val)
		return val
	}
	return defval
}

func getStringFromDB(sqlx string, defval string) string {

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	if rows.Next() {
		var val string
		rows.Scan(&val)
		return val
	}
	return defval
}

func main() {

	fmt.Println(PROGRAMVERSION)
	flag.Parse()

	dbx, _ := filepath.Abs(*DBNAME)
	fmt.Printf("Using %v\n\n", dbx)

	var err error
	DBH, err = sql.Open("sqlite3", dbx)
	checkerr(err)

	http.HandleFunc("/", show_root)
	http.HandleFunc("/rblr", import_rblr)
	err = http.ListenAndServe(":"+*HTTPPort, nil)
	checkerr(err)
}

func show_root(w http.ResponseWriter, r *http.Request) {

	fmt.Println(`no-op called`)
	fmt.Fprintf(w, `<p>no-op called</p><p>%v</p>`, r)
}
