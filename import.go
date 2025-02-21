package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
)

//go:embed rupert.js
var script string

const x = `{"filetype":"iba1000","asat":"2025-02-17T14:07","entrants":[{"EntrantID":14,"EntrantStatus":8,"Rider":{"First":"Kevin","Last":"Ansell","IBA":"","HasIBANumber":false,"RBL":"","Email":"kevin.ansell@icloud.com","Phone":"07753 856425","Address1":"7A Keens Lane","Address2":"","Town":"Chinnor","County":"Oxon","Postcode":"OX39 4PF","Country":"United Kingdom"},"Pillion":{"First":"","Last":"","IBA":"","HasIBANumber":false,"RBL":"","Email":"","Phone":"","Address1":"","Address2":"","Town":"","County":"","Postcode":"","Country":""},"NokName":"Benita Ansell","NokRelation":"Wife","NokPhone":"07955 926285","Bike":"Triumph Tiger 900 GT Pro","BikeReg":"KA59 KEV","Route":"C-SCW","OdoStart":"1","OdoFinish":"1012","OdoCounts":"14","StartTime":"2025-02-17T14:06","FinishTime":"2025-02-17T14:07","FundsRaised":{"EntryDonation":"25","SquiresCheque":"","SquiresCash":"","RBLRAccount":"","JustGivingAmt":"","JustGivingURL":""},"FreeCamping":"14","CertificateAvailable":"Y","CertificateDelivered":"N","Tshirt1":"L","Tshirt2":"no thanks","Patches":1,"EditMode":"","Notes":""}
,{"EntrantID":16,"EntrantStatus":8,"Rider":{"First":"Andy","Last":"Millar","IBA":"","HasIBANumber":false,"RBL":"","Email":"longitude180@hotmail.com","Phone":"07983660086","Address1":"13 Beechtree Ave","Address2":"N/a","Town":"Marlow ","County":"Berkshire","Postcode":"SL73NH","Country":"United Kingdom"},"Pillion":{"First":"","Last":"","IBA":"","HasIBANumber":false,"RBL":"","Email":"","Phone":"","Address1":"","Address2":"","Town":"","County":"","Postcode":"","Country":""},"NokName":"Justine Millar","NokRelation":"Spouse","NokPhone":"+44 7852 406987","Bike":"Harley-Davidson FLSTC","BikeReg":"R942FJL","Route":"C-SCW","OdoStart":"1","OdoFinish":"10","OdoCounts":"16","StartTime":"2025-02-17T14:06","FinishTime":"2025-02-17T14:07","FundsRaised":{"EntryDonation":"50","SquiresCheque":"","SquiresCash":"","RBLRAccount":"","JustGivingAmt":"","JustGivingURL":""},"FreeCamping":"16","CertificateAvailable":"Y","CertificateDelivered":"N","Tshirt1":"no thanks","Tshirt2":"no thanks","Patches":1,"EditMode":"","Notes":""}
,{"EntrantID":37,"EntrantStatus":8,"Rider":{"First":"Paul","Last":"Ball","IBA":"83799","HasIBANumber":true,"RBL":"","Email":"pdb2671@gmail.com","Phone":"07788631957","Address1":"9 Rodney Close","Address2":"Hinckley, Leicestershire","Town":"Hinckley","County":"Leicestershire","Postcode":"LE101PA","Country":"United Kingdom"},"Pillion":{"First":"","Last":"","IBA":"","HasIBANumber":false,"RBL":"","Email":"","Phone":"","Address1":"","Address2":"","Town":"","County":"","Postcode":"","Country":""},"NokName":"Lena Ball","NokRelation":"Wife","NokPhone":"07979231258","Bike":"BMW R1250RT","BikeReg":"T4PSD","Route":"F-500AC","OdoStart":"1","OdoFinish":"506","OdoCounts":"37","StartTime":"2025-02-17T14:06","FinishTime":"2025-02-17T14:07","FundsRaised":{"EntryDonation":"50","SquiresCheque":"","SquiresCash":"","RBLRAccount":"","JustGivingAmt":"","JustGivingURL":""},"FreeCamping":"37","CertificateAvailable":"Y","CertificateDelivered":"N","Tshirt1":"L","Tshirt2":"no thanks","Patches":1,"EditMode":"","Notes":""}
,{"EntrantID":63,"EntrantStatus":8,"Rider":{"First":"Stephen","Last":"Allen","IBA":"","HasIBANumber":false,"RBL":"L","Email":"stevenotwet@gmail.com","Phone":"07795806426","Address1":"9 ","Address2":"Pentire Road","Town":"Torpoint","County":"Cornwall","Postcode":"PL112QZ","Country":"United Kingdom"},"Pillion":{"First":"","Last":"","IBA":"","HasIBANumber":false,"RBL":"","Email":"","Phone":"","Address1":"","Address2":"","Town":"","County":"","Postcode":"","Country":""},"NokName":"Michele Potter","NokRelation":"Girlfriend","NokPhone":"07716146127","Bike":"BMW 1200 GSA","BikeReg":"RV64TUO","Route":"A-NCW","OdoStart":"1","OdoFinish":"10","OdoCounts":"63","StartTime":"2025-02-17T14:06","FinishTime":"2025-02-17T14:07","FundsRaised":{"EntryDonation":"25","SquiresCheque":"","SquiresCash":"","RBLRAccount":"","JustGivingAmt":"","JustGivingURL":""},"FreeCamping":"63","CertificateAvailable":"Y","CertificateDelivered":"N","Tshirt1":"no thanks","Tshirt2":"no thanks","Patches":1,"EditMode":"","Notes":""}
]}`

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

func import_rblr(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("json") == "" {
		load_rblr(w, r)
		return
	}
	saturday := r.FormValue("saturday")
	if saturday == "" {
		fmt.Fprint(w, `{"ok":false,"err":"No Saturday date supplied"}`)
		return
	}
	entrants := parse_rblr(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	for _, e := range entrants {
		fmt.Fprintf(w, `%v : %v, %v<br>`, e.EntrantID, e.Rider.Last, e.Rider.First)
	}
}

func load_rblr(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, `<script>%v</script>`, script)
	fmt.Fprint(w, `<form action="/rblr" method="post" enctype="multipart/form-data" >`)
	fmt.Fprint(w, `<input type="hidden" id="json" name="json">`)

	fmt.Fprint(w, `<fieldset>`)
	fmt.Fprint(w, `<label for="saturday">Date of Saturday (yyyy-mm-dd)</label> `)
	fmt.Fprint(w, `<input type="date" id="saturday" name="saturday">`)
	fmt.Fprint(w, `</fieldset>`)
	fmt.Fprint(w, `<fieldset>`)
	fmt.Fprint(w, `<label for="jsonfile">JSON file to upload</label> `)
	fmt.Fprint(w, `<input id="jsonfile" name="jsonfile" type="file" accept=".json" onchange="enableImportLoad(this)">`)
	fmt.Fprint(w, `</fieldset>`)

	fmt.Fprint(w, `<input type="hidden" id="json" name="json" value="">`)

	fmt.Fprint(w, `<input id="submitbutton" disabled type="submit" value="Submit">`)
	fmt.Fprint(w, `</form>`)

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
