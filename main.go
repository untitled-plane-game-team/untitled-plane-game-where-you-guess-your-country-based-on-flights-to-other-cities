package main

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
	"math/rand"
	"time"
	"encoding/json"
)

type GlobalState struct {
	liveGames    map[int64]GameState
	countryNames map[byte]string
	highscores   []int32
}

type FlightInfo struct {
	unlockCost int32
	paramValue int32
	isUnlocked bool
}

type GameState struct {
	gameId           int64
	prevCountryIds   []byte
	lastTimestamp    int64
	score            int32
	currentCountry   byte
	countriesGuessed []byte
	citiesReceived   []FlightInfo
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

func handleMessage(message string, w http.ResponseWriter, r *http.Request) {
	params := strings.Split(message, "/")

	w.Write([]byte(fmt.Sprintf("\"%s\"\n", message)))
	w.Write([]byte(fmt.Sprintf("%d\n", len(params))))
	if len(params) == 0 { w.Write([]byte("T")) }

	_, err := strconv.ParseInt(params[0], 10, 64)

	if err != nil { return }

	w.Write([]byte(message))
}

func handleNewClient(w http.ResponseWriter, r *http.Request) {
	//get new id
	newId := rand.Int63()

	//make sure that new id is not in the map ( although very unlikely )
	_, ok := globalState.liveGames[newId];
	for ok {
		newId = rand.Int63()
		_, ok = globalState.liveGames[newId];
	}

	var newGameState GameState
	newGameState.lastTimestamp = time.Now().Unix()
	newGameState.score = 0
	newGameState.currentCountry = byte(rand.Int() % 195)
	newGameState.prevCountryIds = nil
	newGameState.citiesReceived = nil
	newGameState.countriesGuessed = nil
	newGameState.gameId = newId

	globalState.liveGames[newId] = newGameState

	bytes, _ := json.Marshal(newGameState)

	w.Write(bytes)

}

func main() {
	globalState.liveGames = make(map[int64]GameState)
	globalState.countryNames = make(map[byte]string)
	globalState.highscores = nil

	http.HandleFunc("/", handleHttp)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

