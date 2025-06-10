package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Test for EntrantStatus
const Finisher = 8
const LateFinisher = 10

// CSV format of Finishers exported from ScoreMaster
type rally_Entrant struct {
	RiderName      string
	PillionName    string
	Bike           string
	Placing        int
	Miles          int
	Points         int
	RiderIBA       int
	PillionIBA     int
	BikeReg        string
	Class          int
	Phone          string
	Email          string
	Postcode       string
	Country        string
	Postal_Address string
	RiderRBL       string
	NoviceRider    string
	PillionRBL     string
	NovicePillion  string
}

type RBLR_Route struct {
	Start    string
	Via      string
	Finish   string
	RideName string
	Miles    int
}

// These codes must match those used by Alys. As of 2024 the routes are bidirectional to minimize
// certificate reprints so the 'via' contents need not be reorganized.
var RBLR_Routes = map[string]RBLR_Route{
	"A-NCW": {"Squires cafe", "Berwick-upon-Tweed, Wick and Fort William", "Squires cafe", "RBLR1000-NC", 1006},
	"B-NAC": {"Squires cafe", "Berwick-upon-Tweed, Wick and Fort William", "Squires cafe", "RBLR1000-NA", 1006},
	"C-SCW": {"Squires cafe", "Bangor, Barnstaple, Andover and Lowestoft", "Squires cafe", "RBLR1000-SC", 1015},
	"D-SAC": {"Squires cafe", "Bangor, Barnstaple, Andover and Lowestoft", "Squires cafe", "RBLR1000-SA", 1015},
	"E-5CW": {"Squires cafe", "Workington, Berwick-upon-Tweed and Beverley", "Squires cafe", "RBLR1000-5C", 504},
	"E-5AC": {"Squires cafe", "Workington, Berwick-upon-Tweed and Beverley", "Squires cafe", "RBLR1000-5A", 504},
}

var RBLR_Lates = map[string]RBLR_Route{
	"A-NCW": {"Squires cafe", "Berwick-upon-Tweed, Wick and Fort William", "Squires cafe", "RBLR1000+NC", 1006},
	"B-NAC": {"Squires cafe", "Berwick-upon-Tweed, Wick and Fort William", "Squires cafe", "RBLR1000+NA", 1006},
	"C-SCW": {"Squires cafe", "Bangor, Barnstaple, Andover and Lowestoft", "Squires cafe", "RBLR1000+SC", 1015},
	"D-SAC": {"Squires cafe", "Bangor, Barnstaple, Andover and Lowestoft", "Squires cafe", "RBLR1000+SA", 1015},
	"E-5CW": {"Squires cafe", "Workington, Berwick-upon-Tweed and Beverley", "Squires cafe", "RBLR1000+5C", 504},
	"E-5AC": {"Squires cafe", "Workington, Berwick-upon-Tweed and Beverley", "Squires cafe", "RBLR1000+5A", 504},
}

type RBLR_Person = struct {
	First        string
	Last         string
	IBA          string
	HasIBANumber bool
	RBL          string
	Email        string
	Phone        string
	Address1     string
	Address2     string
	Town         string
	County       string
	Postcode     string
	Country      string
}

type RBLR_Money = struct {
	EntryDonation string
	SquiresCheque string
	SquiresCash   string
	RBLRAccount   string
	JustGivingAmt string
	JustGivingURL string
}

type RBLR_Entrant = struct {
	EntrantID            int
	EntrantStatus        int
	Rider                RBLR_Person
	Pillion              RBLR_Person
	NokName              string
	NokRelation          string
	NokPhone             string
	Bike                 string
	BikeReg              string
	Route                string
	OdoStart             string
	OdoFinish            string
	OdoCounts            string
	StartTime            string
	FinishTime           string
	FundsRaised          RBLR_Money
	FreeCamping          string
	CertificateAvailable string
	CertificateDelivered string
	Tshirt1              string
	Tshirt2              string
	Patches              int
	EditMode             string
	Notes                string
}

type RBLR_Dataset struct {
	Filetype string
	Asat     string
	Entrants []RBLR_Entrant
}

type RBLR_Params struct {
	Ridedate  string
	EventDesc string
}

type Stats struct {
	NewRiders   int
	NewPillions int
	NewRides    int
	Ncw         int
	Nac         int
	Scw         int
	Sac         int
	Cw5         int
	Ac5         int
}

var loadstats *Stats
var newIBAs []string

// Calculate hours:minutes using start and finish times.
func calc_rblr_ridelength(starttime string, finishtime string) (int, int) {

	const timefmt = "2006-01-02T15:04"

	var timezone *time.Location

	timezone, _ = time.LoadLocation("Europe/London")
	ok := true
	st, err := time.ParseInLocation(timefmt, starttime, timezone)
	if err != nil {
		ok = false
	}
	ft, err := time.ParseInLocation(timefmt, finishtime, timezone)
	if err != nil {
		ok = false
	}

	if !ok {
		return 0, 0
	}

	hrs := int(math.Trunc(ft.Sub(st).Hours()))
	mins := int(math.Trunc(ft.Sub(st).Minutes()))
	mins = mins - (hrs * 60)
	return hrs, mins

}

func import_rally(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("thedata") == "" {
		load_rally(w)
		return
	}

	rallycode := strings.ToUpper(r.FormValue("rallycode"))
	if rallycode == "" {
		fmt.Fprint(w, `<p>No rallycode supplied</p>`)
		return
	}
	if getStringFromDB("SELECT RallyTitle FROM rallies WHERE RallyID='"+rallycode+"'", "") == "" {
		rallydesc := r.FormValue("rallydesc")
		make_new_rally(rallycode, rallydesc)
	}
	yr := r.FormValue("rallyyear")
	if len(yr) > 2 {
		yr = yr[2:]
	}
	entrants := parse_rally(r)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, htmlheader)

	fmt.Fprintf(w, `<h1>Update IBAUK Rides database with %v%v results</h1>`, rallycode, yr)

	var stats Stats
	loadstats = &stats

	DBH.Exec("BEGIN")
	defer DBH.Exec("COMMIT")

	fmt.Fprint(w, `<ul>`)
	for _, e := range entrants {
		fmt.Fprintf(w, `<li>%v`, e.RiderName)
		if e.PillionName != "" {
			fmt.Fprintf(w, ` + %v`, e.PillionName)
		}
		fmt.Fprint(w, `</li>`)
		post_rally_entrant_updates(e, rallycode+yr)
	}
	fmt.Fprint(w, `</ul>`)
	fmt.Fprintf(w, `</p><p><strong>%v</strong> rides added to the database</p>`, stats.NewRides)

	fmt.Fprintf(w, `<p>Number of new riders <strong>%v</strong>, number of new pillions <strong>%v</strong></p>`, stats.NewRiders, stats.NewPillions)

	fmt.Fprint(w, `<p><a href="https://rdb.ironbutt.co.uk">Return to Rides database</a>`)

}
func import_rblr(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("thedata") == "" {
		load_rblr(w)
		return
	}
	var rp RBLR_Params
	rp.Ridedate = r.FormValue("saturday")
	if rp.Ridedate == "" {
		fmt.Fprint(w, `<p>No Saturday date supplied</p>`)
		return
	}
	rp.EventDesc = "RBLR 1000 ('" + rp.Ridedate[2:4] + ")"

	entrants := parse_rblr(r)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, htmlheader)

	fmt.Fprint(w, `<h1>Update IBAUK Rides database from RBLR1000 results</h1>`)

	DBH.Exec("BEGIN")
	defer DBH.Exec("COMMIT")
	var stats Stats

	loadstats = &stats
	fmt.Fprint(w, `<p>`)
	for _, e := range entrants {

		// The file includes Finishers and Late Finishers, 1000 mile routes and 500 mile routes
		// but for now we're only interested in IBA qualified results
		IBAFinisher := e.EntrantStatus == Finisher && RBLR_Routes[e.Route].Miles >= 1000
		if !IBAFinisher {
			continue
		}
		//fmt.Fprintf(w, `%v &nbsp; `, e.Rider.Last+",&nbsp;"+e.Rider.First)
		post_rblr_entrant_updates(e, rp)
	}
	fmt.Fprintf(w, `</p><p><strong>%v</strong> rides added to the database</p>`, stats.NewRides)

	fmt.Fprintf(w, `<p>Number of new riders <strong>%v</strong>, number of new pillions <strong>%v</strong></p>`, stats.NewRiders, stats.NewPillions)
	fmt.Fprintf(w, `<p>NCW: <strong>%v</strong>&nbsp;  NAC: <strong>%v</strong>&nbsp;  SCW: <strong>%v</strong>&nbsp;  SAC: <strong>%v</strong>&nbsp; </p>`, stats.Ncw, stats.Nac, stats.Scw, stats.Sac)

	fmt.Fprint(w, `<p>New IBA members</p><ol>`)
	for _, x := range newIBAs {
		fmt.Fprintf(w, `<li>%v</li>`, x)
	}
	fmt.Fprint(w, `</ol>`)
	fmt.Fprint(w, `<p><a href="https://rdb.ironbutt.co.uk">Return to Rides database</a>`)

}

func load_rally(w http.ResponseWriter) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, htmlheader)

	sqlx := "SELECT RallyID,RallyTitle FROM rallies ORDER BY RallyID"
	options := ""

	rallies, err := DBH.Query(sqlx)
	checkerr(err)
	defer rallies.Close()
	var rally string
	var title string
	for rallies.Next() {
		err = rallies.Scan(&rally, &title)
		checkerr(err)
		opt := fmt.Sprintf(`<option value="%v">%v</option>`, rally, title)
		options += opt
	}
	const rallyopts = "<!-- options -->"
	const minyear = "<!-- min -->"
	const maxyear = "<!-- max -->"
	const curyear = "<!-- value -->"
	yr := time.Now().Year()
	x := strings.Replace(loadrallyform, rallyopts, options, 1)
	x = strings.Replace(x, curyear, strconv.Itoa(yr), 1)
	x = strings.Replace(x, maxyear, strconv.Itoa(yr), 1)
	x = strings.Replace(x, minyear, strconv.Itoa(yr-2), 1)
	fmt.Fprint(w, x)

}
func load_rblr(w http.ResponseWriter) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, htmlheader)

	fmt.Fprint(w, loadrblrform)

}

func parse_rally(r *http.Request) []rally_Entrant {

	cdata := r.FormValue("thedata")
	if cdata == "" {
		return []rally_Entrant{}
	}
	rdr := csv.NewReader(strings.NewReader(cdata))

	recs, err := rdr.ReadAll()
	checkerr(err)

	res := make([]rally_Entrant, 0, len(recs))

	skip := true
	for _, ln := range recs {
		if skip {
			skip = false
			continue
		}
		var re rally_Entrant
		re.RiderName = ln[0]
		re.PillionName = ln[1]
		re.Bike = ln[2]
		re.Placing = intval(ln[3])
		re.Miles = intval(ln[4])
		re.Points = intval(ln[5])
		re.RiderIBA = intval(ln[6])
		re.PillionIBA = intval(ln[7])
		re.BikeReg = ln[8]
		re.Class = intval(ln[9])
		re.Phone = ln[10]
		re.Email = ln[11]
		re.Postcode = ln[12]
		re.Country = ln[13]
		re.Postal_Address = ln[14]
		re.RiderRBL = ln[15]
		re.NoviceRider = ln[16]
		re.NovicePillion = ln[17]
		res = append(res, re)
		fmt.Printf("0=%v, 1=%v, pa={ %v }\n", ln[0], ln[1], re.Postal_Address)

	}

	return res

}
func parse_rblr(r *http.Request) []RBLR_Entrant {

	res := make([]RBLR_Entrant, 0)
	jdata := r.FormValue("thedata")
	if jdata == "" {
		return res
	}
	var rblr RBLR_Dataset
	err := json.Unmarshal([]byte(jdata), &rblr)
	checkerr(err)

	return rblr.Entrants
}

func make_new_rally(code string, desc string) {

	sqlx := "INSERT INTO rallies (RallyID,RallyTitle) VALUES(?,?)"
	stmt, err := DBH.Prepare(sqlx)
	checkerr(err)
	defer stmt.Close()
	_, err = stmt.Exec(code, desc)
	checkerr(err)
}

func post_rally_entrant_updates(e rally_Entrant, rc string) {

	post_rally_person_updates(e, rc, false)
	if e.PillionName != "" {
		post_rally_person_updates(e, rc, true)
	}
}

func post_rally_person_updates(e rally_Entrant, rc string, isPillion bool) {

	var riderid int64
	var bikeid int64

	// Unique IDs in the Rides database are not autogenerated, we must calculate and supply

	var ridername string
	var iba int
	var pn string
	if isPillion {
		ridername = e.PillionName
		iba = e.PillionIBA
		pn = "Y"
	} else {
		ridername = e.RiderName
		iba = e.RiderIBA
		pn = "N"
	}

	ad := time.Now().Format("2006-01-02")
	pa := transform_rally_address(e.Postal_Address)

	if iba > 0 {
		riderid = getIntegerFromDB("SELECT riderid FROM riders WHERE IBA_Number='"+strconv.Itoa(iba)+"'", 0)
	}
	if riderid == 0 {
		riderid = getIntegerFromDB("SELECT riderid FROM riders WHERE Rider_Name='"+ridername+"'", 0)
	}
	if riderid == 0 { // Must create new record
		riderid = getIntegerFromDB("SELECT max(riderid) FROM riders", 0) + 1
		sqlx := "INSERT INTO riders (riderid,Rider_Name,IBA_Number,Postal_Address,Postcode,Country,Email,Phone,IsPillion,DateLastActive)"
		sqlx += "VALUES("
		sqlx += fmt.Sprintf("%v,'%v','%v',%v,'%v','%v','%v','%v','%v','%v'", riderid, ridername, iba, pa, e.Postcode, e.Country, e.Email, e.Phone, pn, ad)
		sqlx += ")"
		newIBAs = append(newIBAs, ridername)
		//fmt.Println(sqlx)
		_, err := DBH.Exec(sqlx)
		checkerr(err)
		if pn == "Y" {
			loadstats.NewPillions++
		} else {
			loadstats.NewRiders++
		}
	} else {
		sqlx := "UPDATE riders SET DateLastActive='" + ad + "',Postal_Address=" + pa + ",Postcode='" + e.Postcode + "',Country='" + e.Country + "',Email='" + e.Email + "',Phone='" + e.Phone + "' WHERE riderid=" + fmt.Sprintf("%v", riderid)
		//fmt.Println(sqlx)
		_, err := DBH.Exec(sqlx)
		checkerr(err)
	}
	bikeid = getIntegerFromDB(fmt.Sprintf("SELECT bikeid FROM bikes WHERE riderid=%v AND Bike='%v' AND (ifnull(Registration,'')='%v' OR ifnull(Registration,'')='')", riderid, e.Bike, e.BikeReg), 0)

	// Switch for bike odo is Y=kms, N=miles
	km := "N"
	// Switch not available in Finisher export from ScoreMaster

	if bikeid == 0 {
		bikeid = getIntegerFromDB("SELECT max(bikeid) FROM bikes", 0) + 1
		sqlx := "INSERT INTO bikes (bikeid,riderid,KmsOdo,Bike,Registration) VALUES(?,?,?,?,?)"
		stmt, err := DBH.Prepare(sqlx)
		checkerr(err)
		defer stmt.Close()
		_, err = stmt.Exec(bikeid, riderid, km, e.Bike, e.BikeReg)
		checkerr(err)
		//fmt.Printf("New bike inserted %v\n", bikeid)
	} else {
		sqlx := "UPDATE bikes SET KmsOdo=?,Registration=? WHERE riderid=? AND bikeid=? AND ifnull(Registration,'')=''"
		stmt, err := DBH.Prepare(sqlx)
		checkerr(err)
		defer stmt.Close()
		_, err = stmt.Exec(km, e.BikeReg, riderid, bikeid)
		checkerr(err)
		//fmt.Printf("Bike %v updated\n", bikeid)
	}

	dupecheck := fmt.Sprintf("SELECT recid FROM rallyresults WHERE riderid=%v AND bikeid=%v AND RallyID='%v'", riderid, bikeid, rc)
	x := getIntegerFromDB(dupecheck, 0)
	if x > 0 {
		//fmt.Println("Ride is duplicated")
		return
	}

	uri := getIntegerFromDB("SELECT max(recid) FROM rallyresults", 0) + 1
	sqlx := "INSERT INTO rallyresults (recid,RallyID,FinishPosition,riderid,bikeid,RallyMiles,RallyPoints,Country)"
	sqlx += "VALUES(?,?,?,?,?,?,?,?)"
	//fmt.Println(sqlx)
	stmt, err := DBH.Prepare(sqlx)
	checkerr(err)
	//fmt.Println("All good")
	defer stmt.Close()
	_, err = stmt.Exec(uri, rc, e.Placing, riderid, bikeid, e.Miles, e.Points, e.Country)
	checkerr(err)
	loadstats.NewRides++

}

// This is where database updates are executed for successful RBLR rides
func post_rblr_entrant_updates(e RBLR_Entrant, rp RBLR_Params) {

	post_rblr_person_updates(e, rp, false)
	if e.Pillion.First != "" || e.Pillion.Last != "" || e.Pillion.IBA != "" {
		post_rblr_person_updates(e, rp, true)
	}

}

func transform_rally_address(address string) string {

	// We want to store the address over multiple lines BUT
	// SQLite can't handle escaped chars directly, only via
	// its concatenate function which isn't handled correctly
	// by Prepare/Exec handlers.

	const nl = " || char(13) || char(10) ||"
	pax := strings.Split(address, " | ")

	pa := ""
	if len(pax) > 1 {
		pa = "'" + address + "'"
		return pa
	}

	for ix := 0; ix < len(pax); ix++ {
		if ix > 0 {
			pa += nl
		}
		pa += "'" + strings.TrimSpace(pax[ix]) + "'"

	}

	return pa

}

func transform_rblr_address(p RBLR_Person) string {

	// We want to store the address over multiple lines BUT
	// SQLite can't handle escaped chars directly, only via
	// its concatenate function which isn't handled correctly
	// by Prepare/Exec handlers.
	const nl = " || char(13) || char(10) ||"
	pa := "'" + safesql(p.Address1) + "'"
	if strings.TrimSpace(p.Address2) != "" {
		pa += nl + "'" + safesql(p.Address2) + "'"
	}
	if strings.TrimSpace(p.Town) != "" {
		pa += nl + "'" + safesql(p.Town) + "'"
	}
	if strings.TrimSpace(p.County) != "" {
		pa += nl + "'" + safesql(p.County) + "'"
	}

	return pa

}

func post_rblr_person_updates(e RBLR_Entrant, rp RBLR_Params, isPillion bool) {

	var riderid int64
	var bikeid int64

	// Unique IDs in the Rides database are not autogenerated, we must calculate and supply

	p := e.Rider
	pn := "N"
	if isPillion {
		p = e.Pillion
		pn = "Y"
	}
	ridername := p.First + " " + p.Last

	//fmt.Println(ridername)
	pa := transform_rblr_address(p)

	if strings.TrimSpace(p.IBA) != "" {
		riderid = getIntegerFromDB("SELECT riderid FROM riders WHERE IBA_Number='"+strings.TrimSpace(p.IBA)+"'", 0)
	}
	if riderid == 0 {
		riderid = getIntegerFromDB("SELECT riderid FROM riders WHERE Rider_Name='"+p.First+" "+p.Last+"'", 0)
	}
	if riderid == 0 { // Must create new record
		riderid = getIntegerFromDB("SELECT max(riderid) FROM riders", 0) + 1
		sqlx := "INSERT INTO riders (riderid,Rider_Name,IBA_Number,Postal_Address,Postcode,Country,Email,Phone,IsPillion,DateLastActive,Address1,Address2,Town,County,Rider_First,Rider_Last)"
		sqlx += "VALUES("
		sqlx += fmt.Sprintf("%v,'%v','%v',%v,'%v','%v','%v','%v','%v','%v','%v','%v','%v','%v','%v','%v'", riderid, ridername, p.IBA, pa, p.Postcode, p.Country, p.Email, p.Phone, pn, rp.Ridedate, safesql(p.Address1), safesql(p.Address2), safesql(p.Town), safesql(p.County), safesql(p.First), safesql(p.Last))
		sqlx += ")"

		IBAFinisher := e.EntrantStatus == Finisher && RBLR_Routes[e.Route].Miles >= 1000

		if IBAFinisher {
			newIBAs = append(newIBAs, ridername)
		}

		//fmt.Println(sqlx)
		_, err := DBH.Exec(sqlx)
		checkerr(err)
		if pn == "Y" {
			loadstats.NewPillions++
		} else {
			loadstats.NewRiders++
		}
	} else {
		sqlx := "UPDATE riders SET DateLastActive='" + rp.Ridedate + "',Postal_Address=" + pa + ",Postcode='" + p.Postcode + "',Country='" + p.Country + "',Email='" + p.Email + "',Phone='" + p.Phone + "', Address1='" + safesql(p.Address1) + "',Address2='" + safesql(p.Address2) + "',Town='" + safesql(p.Town) + "',County='" + safesql(p.County) + "',Rider_First='" + safesql(p.First) + "',Rider_Last='" + safesql(p.Last) + "' WHERE riderid=" + fmt.Sprintf("%v", riderid)
		//fmt.Println(sqlx)
		_, err := DBH.Exec(sqlx)
		checkerr(err)
	}
	bikeid = getIntegerFromDB(fmt.Sprintf("SELECT bikeid FROM bikes WHERE riderid=%v AND Bike='%v' AND (ifnull(Registration,'')='%v' OR ifnull(Registration,'')='')", riderid, e.Bike, e.BikeReg), 0)

	// Switch for bike odo is Y=kms, N=miles
	km := "N"
	if e.OdoCounts == "K" {
		km = "Y"
	}

	if bikeid == 0 {
		bikeid = getIntegerFromDB("SELECT max(bikeid) FROM bikes", 0) + 1
		sqlx := "INSERT INTO bikes (bikeid,riderid,KmsOdo,Bike,Registration) VALUES(?,?,?,?,?)"
		stmt, err := DBH.Prepare(sqlx)
		checkerr(err)
		defer stmt.Close()
		_, err = stmt.Exec(bikeid, riderid, km, e.Bike, e.BikeReg)
		checkerr(err)
		//fmt.Printf("New bike inserted %v\n", bikeid)
	} else {
		sqlx := "UPDATE bikes SET KmsOdo=?,Registration=? WHERE riderid=? AND bikeid=? AND ifnull(Registration,'')=''"
		stmt, err := DBH.Prepare(sqlx)
		checkerr(err)
		defer stmt.Close()
		_, err = stmt.Exec(km, e.BikeReg, riderid, bikeid)
		checkerr(err)
		//fmt.Printf("Bike %v updated\n", bikeid)
	}
	rt, ok := RBLR_Routes[e.Route]
	if !ok {
		rt = RBLR_Routes["A-NCW"]
	}
	if e.EntrantStatus != Finisher {
		rt, ok = RBLR_Lates[e.Route]
		if !ok {
			rt = RBLR_Lates["A-NCW"]
		}

	}

	dupecheck := fmt.Sprintf("SELECT NameOnCertificate FROM rides WHERE riderid=%v AND DateRideStart='%v' AND IBA_Ride='%v'", riderid, rp.Ridedate, rt.RideName)
	x := getStringFromDB(dupecheck, "")
	if x == ridername {
		//fmt.Println("Ride is duplicated")
		return
	}

	uri := getIntegerFromDB("SELECT max(URI) FROM rides", 0) + 1
	rideid := getIntegerFromDB("SELECT recid FROM ridenames WHERE IBA_Ride='"+rt.RideName+"'", 0)
	sqlx := "INSERT INTO rides (URI,riderid,NameOnCertificate,DateRideStart,DateRideFinish,IBA_Ride,IsPillion,EventName,KmsOdo,TotalMiles,bikeid,StartPoint,FinishPoint,MidPoints,DateRcvd,RideVerifier,DateVerified,DateCertSent,IBA_RideID,DatePayRcvd,DatePayReq,ShowRoH,StartOdo,FinishOdo,TimeStart,TimeFinish,RideHours,RideMins,VerifierNotes)"
	sqlx += "VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	//fmt.Println(sqlx)
	stmt, err := DBH.Prepare(sqlx)
	checkerr(err)
	//fmt.Println("All good")
	defer stmt.Close()

	showRoH := "Y"
	IBAFinisher := e.EntrantStatus == Finisher && rt.Miles >= 1000
	if !IBAFinisher {
		showRoH = "N"
	}

	hrs, mins := calc_rblr_ridelength(e.StartTime, e.FinishTime)
	_, err = stmt.Exec(uri, riderid, ridername, rp.Ridedate, rp.Ridedate, rt.RideName, pn, rp.EventDesc, km, rt.Miles, bikeid, rt.Start, rt.Finish, rt.Via, rp.Ridedate, "RBLR", rp.Ridedate, rp.Ridedate, rideid, rp.Ridedate, rp.Ridedate, showRoH, e.OdoStart, e.OdoFinish, e.StartTime, e.FinishTime, hrs, mins, e.Notes)
	checkerr(err)
	loadstats.NewRides++
	switch e.Route {
	case "A-NCW":
		loadstats.Ncw++
	case "B-NAC":
		loadstats.Nac++
	case "C-SCW":
		loadstats.Scw++
	case "D-SAC":
		loadstats.Sac++
	case "E-5CW":
		loadstats.Cw5++
	case "F-5AC":
		loadstats.Ac5++
	}

}

func safesql(x string) string {

	return strings.ReplaceAll(strings.TrimSpace(x), "'", "''")
}
