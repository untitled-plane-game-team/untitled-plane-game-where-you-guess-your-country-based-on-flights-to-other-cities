package main

import (
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
var COST_WRONG_GUESS, COST_INFO_0, COST_INFO_1, COST_INFO_2, COST_INFO_3,
    COST_INFO_4, COST_INFO_5, COST_NEW_FLIGHT, REWARD_CORRECT_GUESS int32

func initCosts() {
	COST_WRONG_GUESS = 1000
	COST_INFO_0 = 0
	COST_INFO_1 = 0
	COST_INFO_2 = 300
	COST_INFO_3 = 300
	COST_INFO_4 = 300
	COST_INFO_5 = 300
	COST_NEW_FLIGHT = 700
	REWARD_CORRECT_GUESS = 26500
}

func getStrippedGameState(gameState GameState) GameState {
	var newGameState GameState

	newGameState.LastTimestamp = gameState.LastTimestamp
	newGameState.Score = gameState.Score
	//newGameState.currentCountry = nil
	// Don't care for this information, won't get encoded anyway
	newGameState.PrevCountryIds = []uint16{}
	newGameState.CitiesReceived = []string{}
	newGameState.FlightData = []FlightInfo{}
	newGameState.CountriesGuessed = []uint16{}
	newGameState.GameId = gameState.GameId

	for _, value := range gameState.PrevCountryIds {
		newGameState.PrevCountryIds = append(newGameState.PrevCountryIds, value)
	}
	for _, value := range gameState.CitiesReceived {
		newGameState.CitiesReceived = append(newGameState.CitiesReceived, value)
	}
	for _, value := range gameState.CountriesGuessed {
		newGameState.CountriesGuessed = append(newGameState.CountriesGuessed, value)
	}
	for _, value := range gameState.FlightData {
		if value.UnlockCost != 0 {
			value.ParamValue = 0
		}
		newGameState.FlightData = append(newGameState.FlightData, value)
	}

	return newGameState
}

func getCountry() uint16 {
	return uint16(rand.Int31() % 195)
}

func getGame(gameState GameState, w http.ResponseWriter) {
	bytes, _ := json.Marshal(getStrippedGameState(gameState))
	w.Write(bytes)
}

func unlockParam(gameState GameState, flightInfoId int32, w http.ResponseWriter) {
	gameState.LastTimestamp = time.Now().Unix()
	gameState.Score -= gameState.FlightData[flightInfoId].UnlockCost
	gameState.FlightData[flightInfoId].UnlockCost = 0

	getGame(gameState, w)
}

func util_inList(el uint16, list []uint16) bool {
	for _, v := range list {
		if v == el {
			return true
		}
	}
	return false
}

func startNewRound(gameState GameState, w http.ResponseWriter) {
	gameState.LastTimestamp = time.Now().Unix()
	gameState.Score += REWARD_CORRECT_GUESS
	gameState.PrevCountryIds = append(gameState.PrevCountryIds, gameState.currentCountry)
	gameState.CountriesGuessed = []uint16{}
	gameState.CitiesReceived = []string{}
	gameState.FlightData = []FlightInfo{}
	gameState.currentCountry = getCountry()


	for util_inList(gameState.currentCountry, gameState.PrevCountryIds)  {
		// while in list of already guessed countried, get a new one
		gameState.currentCountry = getCountry()
	}

	globalState.liveGames[gameState.GameId] = gameState
	getGame(gameState, w)
}

func guessCountry(gameState GameState, countryIndex uint16, w http.ResponseWriter) {
	if countryIndex == gameState.currentCountry {
		startNewRound(gameState, w)
	} else {
		gameState.LastTimestamp = time.Now().Unix()
		gameState.Score -= COST_WRONG_GUESS
		globalState.liveGames[gameState.GameId] = gameState
		getGame(gameState, w)
	}
}

func unlockFlight(gameState GameState, city string, w http.ResponseWriter) {
	gameState.LastTimestamp = time.Now().Unix()
	// TODO : make an actual API call, process that

	var API_cityName string
	var API_cheapestCost, API_cheapestTime, API_cheapestStops, API_averageCost,
	    API_averageTime, API_averageStops int32

	gameState.CitiesReceived = append(gameState.CitiesReceived, API_cityName)
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_0, API_cheapestCost})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_1, API_cheapestTime})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_2, API_cheapestStops})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_3, API_averageCost})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_4, API_averageTime})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_5, API_averageStops})
	gameState.Score -= COST_NEW_FLIGHT

	globalState.liveGames[gameState.GameId] = gameState

	getGame(gameState, w)
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")

	if (len(message) == 0) {
		handleNewClient(w)
	} else {
		handleMessage(message, w)
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
	// Get new id
	newId := rand.Int63()

	// Make sure that new id is not in the map ( although very unlikely )
	_, ok := globalState.liveGames[newId];
	for ok {
		newId = rand.Int63()
		_, ok = globalState.liveGames[newId];
	}

	globalState.liveGames[newId] = GameState{
		LastTimestamp: time.Now().Unix(),
		Score: 0,
		currentCountry: getCountry(),
		PrevCountryIds: []uint16{},
		CitiesReceived: []string{},
		FlightData: []FlightInfo{},
		CountriesGuessed: []uint16{},
		GameId: newId }

	getGame(globalState.liveGames[newId], w)
}

func main() {
	initCosts()
	globalState.liveGames = make(map[int64]GameState)
	globalState.countryNames = make(map[uint16]string)
	globalState.highscores = nil

	http.HandleFunc("/", handleHttp)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

