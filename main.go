package main

import (
	"net/http"
	"strings"
	"math/rand"
)

type GlobalState struct {
	liveThreads map[int64]GameState
	countryNames map[byte]string
	r rand.Rand
	highscores []int32
}

type FlightInfo struct {
	unlock_cost int32
	param_value int32
	is_unlocked bool
}

type GameState struct {
	prev_country_ids  []byte
	last_timestamp    int64
	score             int32
	current_country   byte
	countries_guessed []byte
	cities_received   []FlightInfo
}


func handleHttp(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message

	w.Write([]byte(message))
}

func main() {
	http.HandleFunc("/", handleHttp)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
