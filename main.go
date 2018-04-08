package main

import (
	"net/http"
	"strings"
	"strconv"
	"math/rand"
	"time"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"math"
	"regexp"
	"unsafe"
)

type GlobalState struct {
	liveGames    map[string]GameState
	countryNames map[uint16]string
	countryCodes map[uint16]string
	highscores   []int32
}

type FlightInfo struct {
	UnlockCost int32
	ParamValue int32
}

type GameState struct {
	GameId           string
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

type CountryCodes struct {
	Countries []struct {
		Code string `json:"Code"`
		Name string `json:"Name"`
	} `json:"Countries"`
}

func initStuff() {
	COST_WRONG_GUESS = 1000
	COST_INFO_0 = 0
	COST_INFO_1 = 0
	COST_INFO_2 = 300
	COST_INFO_3 = 300
	COST_INFO_4 = 300
	COST_NEW_FLIGHT = 700
	REWARD_CORRECT_GUESS = 26500

	globalState.liveGames = make(map[string]GameState)
	globalState.countryNames = make(map[uint16]string)
	globalState.countryCodes = make(map[uint16]string)
	globalState.highscores = nil
	COST_INFO_5 = 300
	fileData, _ := ioutil.ReadFile("Countries.json")

	var ccodes CountryCodes
	_ = json.Unmarshal(fileData, &ccodes)

	for _, value := range ccodes.Countries {
		id := uint16(value.Code[0] - 'A') * 16 + uint16(value.Code[1] - 'A')
		globalState.countryNames[id] = value.Name
		globalState.countryCodes[id] = value.Code
	}

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
	target :=  uint16(rand.Int31() % 256)
	for len(globalState.countryNames[target]) < 2 {
		target = uint16(rand.Int31() % 256)
	}
	return target
}

func getGame(gameState GameState, w http.ResponseWriter) {
	bytes, _ := json.Marshal(gameState)//getStrippedGameState(gameState))
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
	//oldCountry := gameState.currentCountry

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


type QuoteResults struct {
	Quotes []struct {
		QuoteID     int  `json:"QuoteId"`
		MinPrice    float64  `json:"MinPrice"`
		Direct      bool `json:"Direct"`
		OutboundLeg struct {
			CarrierIds    []int  `json:"CarrierIds"`
			OriginID      int    `json:"OriginId"`
			DestinationID int    `json:"DestinationId"`
			DepartureDate string `json:"DepartureDate"`
		} `json:"OutboundLeg"`
		QuoteDateTime string `json:"QuoteDateTime"`
	} `json:"Quotes"`
	Places []struct {
		PlaceID        int    `json:"PlaceId"`
		IataCode       string `json:"IataCode"`
		Name           string `json:"Name"`
		Type           string `json:"Type"`
		SkyscannerCode string `json:"SkyscannerCode"`
		CityName       string `json:"CityName"`
		CityID         string `json:"CityId"`
		CountryName    string `json:"CountryName"`
	} `json:"Places"`
	Carriers []struct {
		CarrierID int    `json:"CarrierId"`
		Name      string `json:"Name"`
	} `json:"Carriers"`
	Currencies []struct {
		Code                        string `json:"Code"`
		Symbol                      string `json:"Symbol"`
		ThousandsSeparator          string `json:"ThousandsSeparator"`
		DecimalSeparator            string `json:"DecimalSeparator"`
		SymbolOnLeft                bool   `json:"SymbolOnLeft"`
		SpaceBetweenAmountAndSymbol bool   `json:"SpaceBetweenAmountAndSymbol"`
		RoundingCoefficient         int    `json:"RoundingCoefficient"`
		DecimalDigits               int    `json:"DecimalDigits"`
	} `json:"Currencies"`
}

type GoogleMapsApiResult struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Bounds struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"bounds"`
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}


// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}


func getFlight(countryid uint16, from string, to string, date int64) (string, int32, int32, int32, int32, int32, int32) {
	str_date1 := fmt.Sprintf("%04d-%02d", time.Unix(date, 0).Year(), time.Unix(date, 0).Month())
	str_date2 := fmt.Sprintf("%04d-%02d", time.Unix(date, 0).Year(), time.Unix(date + 2592000, 0).Month())
	apimsg1 := "http://partners.api.skyscanner.net/apiservices/browsequotes/v1.0/DK/EUR/en-US/" + from + "/" + to +"/" + str_date1 + "/?apikey=ha186359580606242767782573426762"//, from, to, time.Unix(date, 0).Format("yyyy-mm"))
	apimsg2 := "http://partners.api.skyscanner.net/apiservices/browsequotes/v1.0/DK/EUR/en-US/" + from + "/" + to +"/" + str_date2 + "/?apikey=ha186359580606242767782573426762"//, from, to, time.Unix(date + 2592000, 0).Format("yyyy-mm"))

	fmt.Println(apimsg1)
	fmt.Println(apimsg2)

	req1, err1 := http.NewRequest("GET", apimsg1, nil)
	req2, err2 := http.NewRequest("GET", apimsg2, nil)

	if err1 != nil || err2 != nil { return "Fuck1", 0, 0, 0, 0, 0 ,0 }

	client1 := &http.Client{}
	client2 := &http.Client{}

	req1.Header.Set("Accept", "application/json")
	req2.Header.Set("Accept", "application/json")


	res1, err1 := client1.Do(req1)
	res2, err2 := client2.Do(req2)

	fmt.Println(res1.StatusCode)
	fmt.Println(res2.StatusCode)

	if err1 != nil {return "Fuck, " + err1.Error(), 0, 0, 0, 0, 0 ,0 }
	if err2 != nil {return "Fuck2, " + err2.Error(), 0, 0, 0, 0, 0, 0}

	defer res1.Body.Close()
	defer res2.Body.Close()

	var quotes1 QuoteResults
	var quotes2 QuoteResults

	json.NewDecoder(res1.Body).Decode(&quotes1)
	json.NewDecoder(res2.Body).Decode(&quotes2)

	var targetPlaceId1 int
	var targetPlaceId2 int
	var cityName string

	for _, value := range quotes1.Places {
		if value.CityID == to {
			//t, _ := strconv.ParseInt(value.PlaceID, 10, 32)
			targetPlaceId1 = value.PlaceID
		}
	}
	for _, value := range quotes2.Places {
		if value.CityID == to {
			//t, _ := strconv.ParseInt(value.CityID)
			targetPlaceId2 = value.PlaceID
			cityName = value.CityName
		}
	}

	type location struct{
		lat float64
		lon float64
	}

	var quotes []struct {
		price float64
		num int32
	}

	targetPlaceId2 = targetPlaceId1
	fmt.Sprintf("%d", targetPlaceId2)

	for _, value := range quotes1.Quotes {
		//if value.OutboundLeg.DestinationID == targetPlaceId1 {
			var num int32
			if value.Direct {
				num = 0
			} else {
				num = 1
			}
			var x struct {
				price float64
				num int32
			}
			x.price = value.MinPrice
			x.num = num

			quotes = append(quotes, x)

		//}
	}

	for _, value := range quotes2.Quotes {
		//if value.OutboundLeg.DestinationID == targetPlaceId2 {
			var num int32
			if value.Direct {
				num = 0
			} else {
				num = 1
			}
			var x struct {
				price float64
				num int32
			}
			x.price = value.MinPrice
			x.num = num

			quotes = append(quotes, x)
		//}
	}


	sort.Slice(quotes, func(i, j int) bool { return quotes[i].price < quotes[j].price })
	median := len(quotes) / 2

	re := regexp.MustCompile("\\s")
	originloc := re.ReplaceAllString(globalState.countryNames[countryid], "%20")

	apiurl := "https://maps.googleapis.com/maps/api/geocode/json?address=\"" + originloc + "\"&key=AIzaSyA6izjWGjbepQaujh4le5VdcO7vMG1NFQo"
	apiurl2 := "https://maps.googleapis.com/maps/api/geocode/json?address=\"" + re.ReplaceAllString(cityName, "%20") + "\"&key=AIzaSyA6izjWGjbepQaujh4le5VdcO7vMG1NFQo"

	//fmt.Println(apiurl)
	//fmt.Println(apiurl2)

	req1, err1 = http.NewRequest("GET", apiurl, nil)
	req2, err2 = http.NewRequest("GET", apiurl2, nil)




	if err1 != nil || err2 != nil { return "", 0, 0, 0, 0, 0 ,0 }

	client1 = &http.Client{}
	client2 = &http.Client{}


	res1, err1 = client1.Do(req1)
	res2, err2 = client2.Do(req2)


	if err1 != nil || err2 != nil { return "", 0, 0, 0, 0, 0 ,0 }

	defer res1.Body.Close()
	defer res2.Body.Close()

	var loc0 GoogleMapsApiResult
	var loc1 GoogleMapsApiResult

	json.NewDecoder(res1.Body).Decode(&loc0)
	json.NewDecoder(res2.Body).Decode(&loc1)

	fmt.Println(loc0)
	fmt.Println(loc1)

	var lat0, lon0, lat1, lon1 float64

	lat0 = loc0.Results[0].Geometry.Location.Lat
	lon0 = loc0.Results[0].Geometry.Location.Lng
	lat1 = loc1.Results[0].Geometry.Location.Lat
	lon1 = loc1.Results[0].Geometry.Location.Lng

	//lat0 = 0
	//lon0 = 0
	//lat1 = 50
	//lon1 = float64(median)


	//return "????", 0, 0, 0, 0, 0, 0

	return cityName, int32(quotes[0].price), int32(Distance(lat0, lon0, lat1, lon1)/1000), int32(quotes[0].num),
		int32(quotes[median].price), int32(Distance(lat0, lon0, lat1, lon1)/1000), int32(quotes[median].num)
		//int32(quotes[0].price), int32(Distance(lat0, lon0, lat1, lon1)/1000), int32(quotes[0].num)
}

func unlockFlight(gameState GameState, city string, w http.ResponseWriter) {
	gameState.LastTimestamp = time.Now().Unix()
	// TODO : make an actual API call, process that

	//countryString := fmt.Sprintf("%c%c", byte(gameState.currentCountry / 16) + 'A', byte(gameState.currentCountry % 16) + 'A')


	API_cityName,
	API_cheapestCost,
	API_cheapestTime,
	API_cheapestStops,
	API_averageCost,
	API_averageTime,
	API_averageStops := getFlight(gameState.currentCountry, globalState.countryCodes[gameState.currentCountry], city, gameState.LastTimestamp)

	gameState.CitiesReceived = append(gameState.CitiesReceived, API_cityName)
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_0, API_cheapestCost})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_1, API_cheapestTime})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_2, API_cheapestStops})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_3, API_averageCost})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_4, API_averageTime})
	gameState.FlightData = append(gameState.FlightData, FlightInfo{COST_INFO_5, API_averageStops})
	gameState.Score -= COST_NEW_FLIGHT

	globalState.liveGames[gameState.GameId] = gameState

	getGame(globalState.liveGames[gameState.GameId], w)
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")

	if (len(message) == 0) {
		handleNewClient(w)
	} else {
		handleMessage(message, w)
	}
}

func handleMessage(message string, w http.ResponseWriter) {
	params := strings.Split(message, "/")

	if len(params) < 2 { return }

	idString, method, params := params[0], params[1], params[2:]


	gameState, ok := globalState.liveGames[idString];
	if ok == false { return }

	switch method {
	case "get_game":
		getGame(gameState, w)

	case "unlock_param":
		if (len(params) == 0) { return }
		flightInfoId, err := strconv.ParseInt(params[0], 10, 32)
		if err != nil { return }
		unlockParam(gameState, int32(flightInfoId), w)

	case "guess_country":
		if (len(params) == 0) { return }
		countryId, err := strconv.ParseInt(params[0], 10, 8)
		if err != nil { return }
		guessCountry(gameState, uint16(countryId), w)

	case "unlock_flight":
		if (len(params) == 0) { return }
		unlockFlight(gameState, params[0], w)
	}
}

func handleNewClient(w http.ResponseWriter) {
	//get new id
	newId := fmt.Sprintf("%d", rand.Int63())


	//make sure that new id is not in the map ( although very unlikely )
	_, ok := globalState.liveGames[newId];
	for ok {
		newId = fmt.Sprintf("%d", rand.Int63())
		_, ok = globalState.liveGames[newId];
	}

	fmt.Println(newId)

	var newGameState GameState
	newGameState.LastTimestamp = time.Now().Unix()
	newGameState.Score = 30000
	newGameState.currentCountry = getCountry()
	newGameState.PrevCountryIds = []uint16{}
	newGameState.CitiesReceived = []string{}
	newGameState.FlightData = []FlightInfo{}
	newGameState.CountriesGuessed = []uint16{}
	newGameState.GameId = newId

	globalState.liveGames[newId] = newGameState

	getGame(globalState.liveGames[newId], w)

}

func main() {
	initStuff()
	var x int
	x = -1
	fmt.Println(unsafe.Sizeof(x))

	http.HandleFunc("/", handleHttp)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

