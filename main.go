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

const PROGRAMVERSION = "Rupert v1.1 Copyright © 2025 Bob Stammers"

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

func intval(x string) int {

	res := 0
	for ix := 0; ix < len(x); ix++ {
		c := x[ix]
		if c < '0' || c > '9' {
			break
		}
		res = res*10 + (int(c) - int('0'))
	}
	return res
}

func main() {

	fmt.Println(PROGRAMVERSION)
	flag.Parse()

	dbx, _ := filepath.Abs(*DBNAME)
	fmt.Printf("Using %v\n", dbx)
	fmt.Printf("Listening on port %v\n\n", *HTTPPort)

	var err error
	DBH, err = sql.Open("sqlite3", dbx)
	checkerr(err)

	http.HandleFunc("/", show_root)
	http.HandleFunc("/help", show_help)
	http.HandleFunc("/rally", import_rally)
	http.HandleFunc("/rblr", import_rblr)
	err = http.ListenAndServe(":"+*HTTPPort, nil)
	checkerr(err)
}

func show_root(w http.ResponseWriter, r *http.Request) {

	show_help(w, r)
	//fmt.Println(`no-op called`)
	//fmt.Fprintf(w, `<p>Rupert no-op (%v) called - maybe return to <a href="https://ironbutt.co.uk">Ironbutt.co.uk</a> or try <a href="/help">help</a></p>`, r.RequestURI)
}
