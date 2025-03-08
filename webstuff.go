package main

import (
	_ "embed"
	"fmt"
	"net/http"
)

//go:embed rupert.js
var script string

//go:embed rupert.css
var css string

var htmlheader = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>Rupert</title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>` + css + `</style>
<script>` + script + `</script>
</head>
<body>
`

var loadrallyform = `
	<h1>Update IBAUK Rides database from rally results</h1>
	<form action="/rally" method="post" enctype="multipart/form-data" >
	<input type="hidden" id="thedata" name="thedata">

	<fieldset>
	<label for="rallyselector">Which rally are you loading?</label>
	<select id="rallyselector" name="rally" onchange="chooseRally(this)">
	<option value="" selected>Rally not yet defined</option>
	<!-- options -->
	</select>
	</fieldset>

	<fieldset>
	<label for="rallycode">Short code of rally</label>
	<input type="text" id="rallycode" name="rallycode" class="rallycode"> 
	<label for="rallydesc">Full rally title</label>
	<input type="text" id="rallydesc" name="rallydesc" class="rallydesc">
	</fieldset>

	<fieldset>
	<label for="rallyyear">Which year's results</label>
	<input type="number" id="rallyyear" name="rallyyear" min="<!-- min -->" max="<!-- max -->" value="<!-- value -->">
	</fieldset>

	<fieldset>
	<label for="thefile">CSV file of results to upload</label> 
	<input id="thefile" name="thefile" type="file" accept=".csv" onchange="enableImportLoad(this)">
	</fieldset>


	<input id="submitbutton" disabled type="submit" value="Submit">
	</form>
`
var loadrblrform = `
	<h1>Update IBAUK Rides database from RBLR1000 results</h1>
	<form action="/rblr" method="post" enctype="multipart/form-data" >
	<input type="hidden" id="thedata" name="thedata">

	<fieldset>
	<label for="saturday">Date of the RBLR Saturday </label> 
	<input type="date" id="saturday" name="saturday">
	</fieldset>
	<fieldset>
	<label for="thefile">JSON file of results to upload</label> 
	<input id="thefile" name="thefile" type="file" accept=".json" onchange="enableImportLoad(this)">
	</fieldset>

	<input id="submitbutton" disabled type="submit" value="Submit">
	</form>
`

var ruperthelp = `
<p>Rupert provides services to the IBAUK Rides database. Services available include:-</p>

<dl>
<dt><a href="/rblr">/rblr</a></dt>
<dd>Update the database with results from the RBLR1000 using the JSON file output from Alys</dd>
</dl>

`

func show_help(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, htmlheader)
	fmt.Fprintf(w, `<p>%v</p>`, PROGRAMVERSION)
	fmt.Fprint(w, ruperthelp)

}
