package main

import (
	_ "embed"
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

var loadrblrform = `
	<h1>Update IBAUK Rides database from RBLR1000 results</h1>
	<form action="/rblr" method="post" enctype="multipart/form-data" >
	<input type="hidden" id="json" name="json">

	<fieldset>
	<label for="saturday">Date of the RBLR Saturday </label> 
	<input type="date" id="saturday" name="saturday">
	</fieldset>
	<fieldset>
	<label for="jsonfile">JSON file of results to upload</label> 
	<input id="jsonfile" name="jsonfile" type="file" accept=".json" onchange="enableImportLoad(this)">
	</fieldset>

	<input type="hidden" id="json" name="json" value="">

	<input id="submitbutton" disabled type="submit" value="Submit">
	</form>
`
