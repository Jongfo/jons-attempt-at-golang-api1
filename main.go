package main

import (
	"io"
	"net/http"
	"os"
)

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello World")
}

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/", hello)
	http.ListenAndServe(":"+port, nil)
	/*
		s := "http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc"
		track, err := igc.ParseLocation(s)
		if err != nil {
			fmt.Errorf("Problem reading the track", err)
		}

		fmt.Printf("Pilot: %s, gliderType: %s, date: %s",
			track.Pilot, track.GliderType, track.Date.String())
	*/
}
