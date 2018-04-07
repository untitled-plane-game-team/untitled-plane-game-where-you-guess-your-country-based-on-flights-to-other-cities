package main

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"math/rand"
)

type GlobalState struct {
	liveGames    map[int64]GameState
	countryNames map[byte]string
	r            rand.Rand
	highscores   []int32
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

var globalState GlobalState

func handleHttp(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")

	if (len(message) == 0) {
		handleNewClient(w, r)
	} else {
		handleMessage(message, w, r)
	}
}

func handleMessage(message string, http.ResponseWriter, r *http.Request) {
	params := strings.Split(message, "/")

	w.Write([]byte(fmt.Sprintf("\"%s\"\n", message)))
	w.Write([]byte(fmt.Sprintf("%d\n", len(params))))
	if len(params) == 0 { w.Write([]byte("T")) }

	_, err := strconv.ParseInt(params[0], 10, 64)

	if err != nil { return }

	w.Write([]byte(message))
}

func main() {
	globalState.liveGames = make(map[int64]GameState)
	globalState.countryNames = make(map[byte]string)
	http.HandleFunc("/", handleHttp)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

