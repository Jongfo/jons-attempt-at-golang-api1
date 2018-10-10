package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	igc "github.com/marni/goigc"
)

//Service contains data about our service
type Service struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

func handl404(w http.ResponseWriter, r *http.Request) {
	//TODO: set header to 404
	w.Header().Set("header_name", "404")
	fmt.Fprint(w, "Error 404")
}

func handlAPI(w http.ResponseWriter, r *http.Request) {

}

/* test func that helped me understand task better
func hello(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "Hello World")
	fmt.Fprint(w, "Hello Person!")
}*/

func mariusz(w http.ResponseWriter, r *http.Request) {
	s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	track, err := igc.ParseLocation(s)
	if err != nil {
		log.Fatal("Problem reading the track", err)
	}

	fmt.Fprintf(w, "Pilot: %s, gliderType: %s, date: %s",
		track.Pilot, track.GliderType, track.Date.String())
}

func main() { /*
		port := os.Getenv("PORT")
		http.HandleFunc("/", mariusz)
		http.HandleFunc("/hello", hello)
		http.ListenAndServe(":"+port, nil)*/

	r := mux.NewRouter()
	r.HandleFunc("/", handl404)
	r.HandleFunc("/igcinfo/api", handlAPI)
	http.Handle("/", r)
	http.ListenAndServe(":", nil)
}
