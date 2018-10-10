package main

import (
	"fmt"
	"net/http"
	"os"

	igc "github.com/marni/goigc"
)

//test func that helped me understand task better
func hello(w http.ResponseWriter, r *http.Request) {
	//io.WriteString(w, "Hello World")
	fmt.Fprint(w, "Hello Person!")
}

func mariusz(w http.ResponseWriter, r *http.Request) {
	s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
	track, err := igc.ParseLocation(s)
	if err != nil {
		fmt.Errorf("Problem reading the track", err)
	}

	fmt.Fprintf(w, "Pilot: %s, gliderType: %s, date: %s",
		track.Pilot, track.GliderType, track.Date.String())

}

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/", mariusz)
	http.ListenAndServe(":"+port, nil)

}
