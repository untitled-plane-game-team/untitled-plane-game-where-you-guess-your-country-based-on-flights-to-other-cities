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
	countryNames map[uint16]string
	highscores   []int32
}

type FlightInfo struct {
	UnlockCost int32
	ParamValue int32
}

type GameState struct {
	GameId           int64
	PrevCountryIds   []uint16
	LastTimestamp    int64
	Score            int32
	currentCountry   uint16
	CountriesGuessed []uint16
	CitiesReceived   []string
	FlightData       []FlightInfo
}

var globalState GlobalState

func getGame(GameState gameState, w http.ResponseWriter) {

}

func unlockParam(GameState gameState, int32 flightInfoId, w http.ResponseWriter) {

}

func guessCountry(GameState gameState, byte countryIndex) {

}

func unlockFlight(GameState gameState, string city) {

}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")

	if (len(message) == 0) {
		handleNewClient(w, r)
	} else {
		handleMessage(message, w, r)
	}
}

func handleMessage(message string, w http.ResponseWriter) {
	if len(params) < 2 { return }

	params := strings.Split(message, "/")
	idString, method, params := params[0], params[1], params[2:]
	gameId, err := strconv.ParseInt(idString, 10, 64)

	if err != nil { return }
	if gameState, err := globalState.liveGames[gameId]; err != nil { return }

	switch method {
	case "get_game":
		getGame(gameState, w)
	case "unlock_param":
		if (len(params) == 0) { return }
		flightInfoId, err := strconv.parseInt(params[0], 10, 32)
		if err != nil { return }
		unlockParam(gameState, , w)
	case "guess_country":
		if (len(params) == 0) { return }
		countryId, err := strconv.parseInt(params[0], 10, 8)
		if err != nil { return }
		guessCountry(gameState, countryId, w)
	case "unlock_flight":
		if (len(params) == 0) { return }
		unlockFlight(gameState, params[0], w)
	}
}

func handleNewClient(w http.ResponseWriter) {
	//get new id
	newId := rand.Int63()

	//make sure that new id is not in the map ( although very unlikely )
	_, ok := globalState.liveGames[newId];
	for ok {
		newId = rand.Int63()
		_, ok = globalState.liveGames[newId];
	}

	var newGameState GameState
	newGameState.LastTimestamp = time.Now().Unix()
	newGameState.Score = 0
	newGameState.currentCountry = uint16(rand.Int31() % 195)
	newGameState.PrevCountryIds = []uint16{}
	newGameState.CitiesReceived = []FlightInfo{}
	newGameState.CountriesGuessed = []uint16{}
	newGameState.GameId = newId

	globalState.liveGames[newId] = newGameState

	bytes, _ := json.Marshal(newGameState)

	w.Write(bytes)

}

func main() {
	globalState.liveGames = make(map[int64]GameState)
	globalState.countryNames = make(map[uint16]string)
	globalState.highscores = nil

	http.HandleFunc("/", handleHttp)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

