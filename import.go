package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

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

type RBLR_Stats struct {
	NewRiders   int
	NewPillions int
	NewRides    int
	Ncw         int
	Nac         int
	Scw         int
	Sac         int
}

var loadstats *RBLR_Stats

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

func import_rblr(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("json") == "" {
		load_rblr(w, r)
		return
	}
	var rp RBLR_Params
	rp.Ridedate = r.FormValue("saturday")
	if rp.Ridedate == "" {
		fmt.Fprint(w, `{"ok":false,"err":"No Saturday date supplied"}`)
		return
	}
	rp.EventDesc = "RBLR 1000('" + rp.Ridedate[2:4] + ")"
	entrants := parse_rblr(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, htmlheader)

	fmt.Fprint(w, `<h1>Update IBAUK Rides database from RBLR1000 results</h1>`)

	DBH.Exec("BEGIN")
	defer DBH.Exec("COMMIT")
	var stats RBLR_Stats

	loadstats = &stats
	fmt.Fprint(w, `<p>`)
	for _, e := range entrants {
		fmt.Fprintf(w, `%v &nbsp; `, e.Rider.Last+",&nbsp;"+e.Rider.First)
		post_rblr_entrant_updates(e, rp)
	}
	fmt.Fprintf(w, `</p><p><strong>%v</strong> rides added to the database</p>`, stats.NewRides)

	fmt.Fprintf(w, `<p>Number of new riders <strong>%v</strong>, number of new pillions <strong>%v</strong></p>`, stats.NewRiders, stats.NewPillions)
	fmt.Fprintf(w, `<p>NCW: <strong>%v</strong>&nbsp;  NAC: <strong>%v</strong>&nbsp;  SCW: <strong>%v</strong>&nbsp;  SAC: <strong>%v</strong>&nbsp;</p>`, stats.Ncw, stats.Nac, stats.Scw, stats.Sac)
}

func load_rblr(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprint(w, htmlheader)

	fmt.Fprint(w, loadrblrform)

}
func parse_rblr(r *http.Request) []RBLR_Entrant {

	res := make([]RBLR_Entrant, 0)
	jdata := r.FormValue("json")
	if jdata == "" {
		return res
	}
	var rblr RBLR_Dataset
	err := json.Unmarshal([]byte(jdata), &rblr)
	checkerr(err)

	return rblr.Entrants
}

// This is where database updates are executed for successful RBLR rides
func post_rblr_entrant_updates(e RBLR_Entrant, rp RBLR_Params) {

	post_rblr_person_updates(e, rp, false)
	if e.Pillion.First != "" || e.Pillion.Last != "" || e.Pillion.IBA != "" {
		post_rblr_person_updates(e, rp, true)
	}

}

func transform_rblr_address(p RBLR_Person) string {

	// We want to store the address over multiple lines BUT
	// SQLite can't handle escaped chars directly, only via
	// its concatenate function which isn't handled correctly
	// by Prepare/Exec handlers.
	const nl = " || char(13) || char(10) ||"
	pa := "'" + strings.TrimSpace(p.Address1) + "'"
	if strings.TrimSpace(p.Address2) != "" {
		pa += nl + "'" + strings.TrimSpace(p.Address2) + "'"
	}
	if strings.TrimSpace(p.Town) != "" {
		pa += nl + "'" + strings.TrimSpace(p.Town) + "'"
	}
	if strings.TrimSpace(p.County) != "" {
		pa += nl + "'" + strings.TrimSpace(p.County) + "'"
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

	pa := transform_rblr_address(p)

	if strings.TrimSpace(p.IBA) != "" {
		riderid = getIntegerFromDB("SELECT riderid FROM riders WHERE IBA_Number='"+strings.TrimSpace(p.IBA)+"'", 0)
	}
	if riderid == 0 {
		riderid = getIntegerFromDB("SELECT riderid FROM riders WHERE Rider_Name='"+p.First+" "+p.Last+"'", 0)
	}
	if riderid == 0 { // Must create new record
		riderid = getIntegerFromDB("SELECT max(riderid) FROM riders", 0) + 1
		sqlx := "INSERT INTO riders (riderid,Rider_Name,IBA_Number,Postal_Address,Postcode,Country,Email,Phone,IsPillion,DateLastActive)"
		sqlx += "VALUES("
		sqlx += fmt.Sprintf("%v,'%v','%v',%v,'%v','%v','%v','%v','%v','%v'", riderid, ridername, p.IBA, pa, p.Postcode, p.Country, p.Email, p.Phone, pn, rp.Ridedate)
		sqlx += ")"

		//fmt.Println(sqlx)
		_, err := DBH.Exec(sqlx)
		checkerr(err)
		if pn == "Y" {
			loadstats.NewPillions++
		} else {
			loadstats.NewRiders++
		}
	} else {
		sqlx := "UPDATE riders SET DateLastActive='" + rp.Ridedate + "',Postal_Address=" + pa + ",Postcode='" + p.Postcode + "',Country='" + p.Country + "',Email='" + p.Email + "',Phone='" + p.Phone + "' WHERE riderid=" + fmt.Sprintf("%v", riderid)
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
		rt, _ = RBLR_Routes["A-NCW"]
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
	hrs, mins := calc_rblr_ridelength(e.StartTime, e.FinishTime)
	_, err = stmt.Exec(uri, riderid, ridername, rp.Ridedate, rp.Ridedate, rt.RideName, pn, rp.EventDesc, km, rt.Miles, bikeid, rt.Start, rt.Finish, rt.Via, rp.Ridedate, "RBLR", rp.Ridedate, rp.Ridedate, rideid, rp.Ridedate, rp.Ridedate, "Y", e.OdoStart, e.OdoFinish, e.StartTime, e.FinishTime, hrs, mins, e.Notes)
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
	}

}
