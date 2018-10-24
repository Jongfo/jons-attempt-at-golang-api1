package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	igc "github.com/marni/goigc"
)

var startTime time.Time

// var registeredTrackIDs []string //a little bit of duplicate data.
var registeredTracks []igc.Track

//Service contains data about our service
type Service struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

//PostURL holds data on POST request.
type PostURL struct {
	URL string `json:"url"`
}

//POSTid contains ID based on PostURL
type POSTid struct {
	ID string `json:"id"`
}

//IDdata contains data on given track id
type IDdata struct {
	Hdate       time.Time `json:"H_date"`       //<date from File Header, H-record>,
	Pilot       string    `json:"pilot"`        //<pilot>,
	Glider      string    `json:"glider"`       //<glider>,
	GliderID    string    `json:"glider_id"`    //<glider_id>,
	TrackLength float64   `json:"track_length"` //<calculated total track length>
}

//errorHandler is a simple self made function to deal with bad requests
func errorHandler(w http.ResponseWriter, code int, mes string) {
	w.WriteHeader(code)
	http.Error(w, http.StatusText(code), code)
	fmt.Fprint(w, mes)
	log.Print(mes)
}

//trackDistance calculates rough distance
func trackDistance(t igc.Track) float64 {
	totalDistance := 0.0
	for j := 0; j < len(t.Points)-1; j++ {
		totalDistance += t.Points[j].Distance(t.Points[j+1])
	}
	return totalDistance
}

//copied code from stackoverflow. Could be improved on.
func diff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

//first self made function. Not relevant to task anymore
func handl404(w http.ResponseWriter, r *http.Request) {
	//sets header to 404
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "We found nothing exept this 404")
}

//writes meta information about our api
func handlAPI(w http.ResponseWriter, r *http.Request) {
	//make a timestamp, and compare it to startTime
	var tim time.Time
	tim = time.Now()
	y, mo, d, h, mi, s := diff(startTime, tim)
	//save the result as a string in the ISO8601 format.
	tim2 := fmt.Sprintf("P%dY%dM%dDT%dH%dM%dS",
		y,  //year
		mo, //month
		d,  //day
		h,  //hour
		mi, //min
		s)  //sec

	//make a json with metadata
	serv := Service{tim2, "Service for IGC tracks.", "v1"}
	js, err := json.Marshal(serv)
	if err != nil {
		str := fmt.Sprintf("Error Marshal: %s", err)
		errorHandler(w, http.StatusInternalServerError, str)
	} else {
		//forward json to requestee
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func handlAPIigc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//check if we have any data yet.
		if len(registeredTracks) == 0 {
			errorHandler(w, http.StatusNoContent, "No tracks registered yet.")
			return
		}
		//store all IDs in a list
		var ids []string
		for i := 0; i < len(registeredTracks); i++ {
			ids = append(ids, registeredTracks[i].UniqueID)
		}
		//turn list into json
		js, err := json.Marshal(ids)
		if err != nil {
			str := fmt.Sprintf("Mershal error: %s", err)
			errorHandler(w, http.StatusInternalServerError, str)
		} else {
			//give json to customer
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	case "POST":
		//make decoder for our POST body
		decoder := json.NewDecoder(r.Body)
		var url PostURL

		//decode the json we've recieved
		err1 := decoder.Decode(&url)
		if err1 != nil {
			str := fmt.Sprintf("Decode error: %s", err1)
			errorHandler(w, http.StatusInternalServerError, str)
			return //something went wrong
		}

		//get track information from provided url.
		track, err2 := igc.ParseLocation(url.URL)
		if err2 != nil {
			str := fmt.Sprintf("Problem reading the track: %s", err2)
			errorHandler(w, http.StatusInternalServerError, str)
			return //something went wrong
		}

		//put the ID of the track we found in a struct to be returned later.
		id := POSTid{track.UniqueID}
		//and turn it into a json
		js, err3 := json.Marshal(id)
		if err3 != nil {
			str := fmt.Sprintf("Marshal error: %s", err3)
			errorHandler(w, http.StatusInternalServerError, str)
			return //something went wrong
		}

		//check if we already have the track registered
		for i := 0; i < len(registeredTracks); i++ {
			if registeredTracks[i].UniqueID == track.UniqueID {
				str := fmt.Sprintf("Error: Already registered")
				errorHandler(w, http.StatusBadRequest, str)
				return //duplicate found
			}
		}

		//adds track track to global slice
		//registeredTrackIDs = append(registeredTrackIDs, track.UniqueID)
		registeredTracks = append(registeredTracks, track)

		//we did everything correctly, hopefully
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)

	default:
		//unexpected request
		str := fmt.Sprintf("Sorry, only GET and POST methods are supported.")
		errorHandler(w, http.StatusBadRequest, str)
	}
}

//writes information about a track on a given ID
func handlAPIigcID(w http.ResponseWriter, r *http.Request) {
	//container for the http adress vaiable {ID}
	vars := mux.Vars(r)

	//goes throug and looks for matching ID
	for i := 0; i < len(registeredTracks); i++ {
		if registeredTracks[i].UniqueID == vars["ID"] {

			//calculates rough distance
			totalDistance := trackDistance(registeredTracks[i])

			//struct to be marshaled and sendt as json
			data := IDdata{
				registeredTracks[i].Date,       //Date from File Header, H-record
				registeredTracks[i].Pilot,      //Pilot name
				registeredTracks[i].GliderType, //Glider type
				registeredTracks[i].GliderID,   //Glider ID
				totalDistance,                  //Calculated total track length
			}

			//make the json struct
			js, err := json.Marshal(data)
			if err != nil {
				str := fmt.Sprintf("Marshal error: %s", err)
				errorHandler(w, http.StatusInternalServerError, str)
			} else {
				//We successfully found data and made the json.
				w.Header().Set("Content-Type", "application/json")
				w.Write(js)
			}
			return //end function after finding track
		}
	}

	//in case we didn't find the track
	str := fmt.Sprintf("Error: Did not find track")
	errorHandler(w, http.StatusBadRequest, str)

}

//writes content of a specified ID and field
func handlAPIigcIDfield(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	//container for the http adress vaiables {ID} and {field}
	vars := mux.Vars(r)

	//Look for matching track ID
	for i := 0; i < len(registeredTracks); i++ {
		if registeredTracks[i].UniqueID == vars["ID"] {
			//Look for matching data name. If found; print data
			switch vars["field"] {
			case "pilot":
				fmt.Fprint(w, registeredTracks[i].Pilot)
			case "glider":
				fmt.Fprint(w, registeredTracks[i].GliderType)
			case "glider_id":
				fmt.Fprint(w, registeredTracks[i].GliderID)
			case "track_length":
				fmt.Fprint(w, trackDistance(registeredTracks[i]))
			case "H_date":
				fmt.Fprint(w, registeredTracks[i].Date)
			default:
				//last field does not match or not implemented yet.
				w.WriteHeader(http.StatusNotFound)
			}
			return
		}
	}
	//we did not find any matches
	w.WriteHeader(http.StatusNotFound)
}

//mariusz is a slightly modified version of the code found in the readme of github.com/marni/goigc
func mariusz(w http.ResponseWriter, r *http.Request) {
	s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	track, err := igc.ParseLocation(s)
	if err != nil {
		str := fmt.Sprintf("Problem reading the track: %s", err)
		errorHandler(w, http.StatusInternalServerError, str)
	} else {
		fmt.Fprintf(w, "Pilot: %s, gliderType: %s, date: %s",
			track.Pilot, track.GliderType, track.Date.String())
	}
}

func init() {
	//make a timestamp for uptime.
	startTime = time.Now()
}

func main() {
	//find our port
	port := os.Getenv("PORT")

	//all our paths using gorilla mux
	r := mux.NewRouter()
	r.HandleFunc("/", handl404)
	r.HandleFunc("/igcinfo/api", handlAPI)
	r.HandleFunc("/igcinfo/api/igc", handlAPIigc)
	r.HandleFunc("/igcinfo/api/igc/{ID}", handlAPIigcID)
	r.HandleFunc("/igcinfo/api/igc/{ID}/{field}", handlAPIigcIDfield)

	//serve our functionallity
	http.Handle("/", r)
	http.ListenAndServe(":"+port, nil)
}
