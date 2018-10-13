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
var registeredTrackIDs []string //a little bit of duplicate data.
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

func uptime() time.Duration {
	return time.Since(startTime)
}

func errorHandler(w http.ResponseWriter, code int, mes string) {
	w.WriteHeader(code)
	http.Error(w, http.StatusText(code), code)
	fmt.Fprint(w, mes)
	log.Print(mes)
}

func handl404(w http.ResponseWriter, r *http.Request) {
	//sets header to 404
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "We found nothing exept this 404 and time: %s", uptime())
}

func handlAPI(w http.ResponseWriter, r *http.Request) {
	t := uptime()
	ti := fmt.Sprintf("P%dY%dM%dDT%dH%dM%dS",
		int(t.Seconds()/31556926),   //year
		int(t.Seconds()/2629743)%12, //month
		int(t.Seconds()/86400)%30,   //day
		int(t.Seconds()/3600)%24,    //hour
		int(t.Seconds()/60)%60,      //min
		int(t.Seconds())%60)         //sec

	serv := Service{ti, "Service for IGC tracks.", "v1"}
	js, err := json.Marshal(serv)
	if err != nil {
		str := fmt.Sprintf("Error Marshal: %s", err)
		errorHandler(w, http.StatusInternalServerError, str)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func handlAPIigc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//TODO:	Task 3
		//w.WriteHeader(http.StatusNotImplemented)
		//fmt.Fprint(w, "Only POST requests works for now.")
		if len(registeredTracks) == 0 {
			errorHandler(w, http.StatusNoContent, "No tracks registered yet.")
			return
		}
		js, err := json.Marshal(registeredTrackIDs)
		if err != nil {
			str := fmt.Sprintf("Mershal error: %s", err)
			errorHandler(w, http.StatusInternalServerError, str)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var url PostURL

		err1 := decoder.Decode(&url)
		if err1 != nil {
			str := fmt.Sprintf("Decode error: %s", err1)
			errorHandler(w, http.StatusInternalServerError, str)
		}

		track, err2 := igc.ParseLocation(url.URL)
		if err2 != nil {
			str := fmt.Sprintf("Problem reading the track: %s", err2)
			errorHandler(w, http.StatusInternalServerError, str)
			return
		}

		id := POSTid{track.UniqueID}

		js, err3 := json.Marshal(id)
		if err3 != nil {
			str := fmt.Sprintf("Marshal error: %s", err3)
			errorHandler(w, http.StatusInternalServerError, str)
		}
		err4 := false
		for i := 0; i < len(registeredTrackIDs); i++ {
			if registeredTrackIDs[i] == track.UniqueID {
				str := fmt.Sprintf("Error: Already registered")
				errorHandler(w, http.StatusBadRequest, str)
				err4 = true
				return
				//break
			}
		}

		if err1 == nil && err2 == nil && err3 == nil && !err4 {

			registeredTrackIDs = append(registeredTrackIDs, track.UniqueID)
			registeredTracks = append(registeredTracks, track)

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}

	default:
		str := fmt.Sprintf("Sorry, only GET and POST methods are supported.")
		errorHandler(w, http.StatusBadRequest, str)
	}
}

func handlAPIigcID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	found := false
	for i := 0; i < len(registeredTrackIDs); i++ {
		if registeredTrackIDs[i] == vars["ID"] {
			data := IDdata{
				registeredTracks[i].Date,
				registeredTracks[i].Pilot,
				registeredTracks[i].GliderType,
				registeredTracks[i].GliderID,
				registeredTracks[i].Task.Distance(),
			}

			js, err := json.Marshal(data)
			if err != nil {
				str := fmt.Sprintf("Marshal error: %s", err)
				errorHandler(w, http.StatusInternalServerError, str)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(js)
				found = true
				//break
			}
			return
		}
	}
	if !found {
		str := fmt.Sprintf("Error: Did not find track")
		errorHandler(w, http.StatusBadRequest, str)
	}

}

func handlAPIigcIDfield(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	vars := mux.Vars(r)
	for i := 0; i < len(registeredTrackIDs); i++ {
		if registeredTrackIDs[i] == vars["ID"] {

			switch vars["field"] {
			case "pilot":
				fmt.Fprint(w, registeredTracks[i].Pilot)
			case "glider":
				fmt.Fprint(w, registeredTracks[i].GliderType)
			case "glider_id":
				fmt.Fprint(w, registeredTracks[i].GliderID)
			case "track_legth":
				fmt.Fprint(w, registeredTracks[i].Task.Distance())
			case "H_date":
				fmt.Fprint(w, registeredTracks[i].Date)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
			//break
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func mariusz(w http.ResponseWriter, r *http.Request) {
	s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	track, err := igc.ParseLocation(s)
	if err != nil {
		str := fmt.Sprintf("Problem reading the track: %s", err)
		errorHandler(w, http.StatusInternalServerError, str)
	}

	fmt.Fprintf(w, "Pilot: %s, gliderType: %s, date: %s",
		track.Pilot, track.GliderType, track.Date.String())
}

func init() {
	startTime = time.Now()
}

func main() {
	port := os.Getenv("PORT")
	r := mux.NewRouter()
	r.HandleFunc("/", handl404)
	r.HandleFunc("/igcinfo/api", handlAPI)
	//	r.HandleFunc("/igcinfo/api/", handlAPI)
	r.HandleFunc("/igcinfo/api/igc", handlAPIigc)
	//	r.HandleFunc("/igcinfo/api/igc/", handlAPIigc)
	r.HandleFunc("/igcinfo/api/igc/{ID}", handlAPIigcID)
	//	r.HandleFunc("/igcinfo/api/igc/{ID}/", handlAPIigcID)
	r.HandleFunc("/igcinfo/api/igc/{ID}/{field}", handlAPIigcIDfield)
	//	r.HandleFunc("/igcinfo/api/igc/{ID}/{field}/", handlAPIigcIDfield)

	http.Handle("/", r)
	http.ListenAndServe(":"+port, nil)
}
